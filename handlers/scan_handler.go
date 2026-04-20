package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ecoscan-ai/cache"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type ScanResult struct {
	Item       string `json:"item"`
	Material   string `json:"material"`
	Disposal   string `json:"disposal"`
	CarbonSave string `json:"carbon_save"`
	Tips       string `json:"tips"`
}

type ScanHandler struct {
	client   *genai.Client
	cache    *cache.Cache
	logger   *slog.Logger
	model    string
	azure    AzureConfig
	provider string
}

type AzureConfig struct {
	Endpoint   string
	APIKey     string
	Deployment string
	APIVersion string
}

func (a AzureConfig) enabled() bool {
	return a.Endpoint != "" && a.APIKey != "" && a.Deployment != ""
}

func NewScanHandler(client *genai.Client, cache *cache.Cache, model string, provider string, azure AzureConfig) *ScanHandler {
	if model == "" {
		model = "gemini-3-flash-preview"
	}
	if provider == "" {
		provider = "gemini"
	}

	return &ScanHandler{
		client:   client,
		cache:    cache,
		logger:   slog.Default(),
		model:    model,
		azure:    azure,
		provider: provider,
	}
}

func (h *ScanHandler) Scan(c *gin.Context) {
	start := time.Now()

	file, err := c.FormFile("image")
	if err != nil {
		h.logger.Warn("No image uploaded", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Please upload an image."})
		return
	}

	src, err := file.Open()
	if err != nil {
		h.logger.Error("Cannot open file", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot open file"})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		h.logger.Error("Cannot read file", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot read file"})
		return
	}

	// Generate hash for caching
	hash := sha256.Sum256(fileBytes)
	hashKey := hex.EncodeToString(hash[:])

	// Check cache first
	if cached, found := h.cache.Get(hashKey); found {
		h.logger.Info("Cache hit", "key", hashKey[:8])
		c.JSON(http.StatusOK, gin.H{
			"result":     cached,
			"source":     "cache",
			"latency_ms": time.Since(start).Milliseconds(),
		})
		return
	}

	result, source, err := h.analyze(fileBytes, contentType)
	if err != nil {
		h.logger.Error("AI analysis failed", "error", err.Error())
		status := http.StatusBadGateway
		errText := err.Error()
		if strings.Contains(errText, "429") || strings.Contains(strings.ToLower(errText), "quota") {
			status = http.StatusTooManyRequests
		}

		c.JSON(status, gin.H{
			"error":    "AI analysis failed",
			"provider": h.provider,
			"detail":   truncateError(errText, 700),
		})
		return
	}

	// Cache the result
	h.cache.Set(hashKey, result)

	h.logger.Info("Scan completed",
		"item", result.Item,
		"disposal", result.Disposal,
		"source", source,
		"latency_ms", time.Since(start).Milliseconds(),
	)

	c.JSON(http.StatusOK, gin.H{
		"result":     result,
		"source":     source,
		"latency_ms": time.Since(start).Milliseconds(),
	})
}

func (h *ScanHandler) analyze(fileBytes []byte, contentType string) (ScanResult, string, error) {
	if h.provider == "azure" {
		result, err := h.analyzeWithAzure(fileBytes, contentType)
		if err != nil {
			return ScanResult{}, "", fmt.Errorf("provider=azure endpoint=%s deployment=%s: %w", normalizeAzureEndpoint(h.azure.Endpoint), h.azure.Deployment, err)
		}
		return result, "azure", nil
	}

	if h.client != nil {
		result, err := h.analyzeWithGemini(fileBytes, contentType)
		if err == nil {
			return result, "gemini", nil
		}
		if h.azure.enabled() {
			h.logger.Warn("Gemini analysis failed, trying Azure fallback", "error", err.Error())
			azureResult, aErr := h.analyzeWithAzure(fileBytes, contentType)
			if aErr == nil {
				return azureResult, "azure-fallback", nil
			}
			return ScanResult{}, "", fmt.Errorf("gemini error: %w; azure fallback error: %v", err, aErr)
		}
		return ScanResult{}, "", err
	}

	if h.azure.enabled() {
		result, err := h.analyzeWithAzure(fileBytes, contentType)
		if err == nil {
			return result, "azure", nil
		}
		return ScanResult{}, "", fmt.Errorf("provider=azure endpoint=%s deployment=%s: %w", normalizeAzureEndpoint(h.azure.Endpoint), h.azure.Deployment, err)
	}

	return ScanResult{}, "", fmt.Errorf("no AI provider configured")
}

func (h *ScanHandler) analyzeWithGemini(fileBytes []byte, contentType string) (ScanResult, error) {
	if h.client == nil {
		return ScanResult{}, fmt.Errorf("gemini client is not configured")
	}

	// Call Gemini with multimodal content.
	contents := []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromBytes(fileBytes, contentType),
			genai.NewPartFromText(`Analyze this item for Earth Day sustainability.
Return JSON with these exact keys:
- item: What is this object?
- material: What material is it made of?
- disposal: One of [Recycle, Compost, Landfill, Hazardous, Reuse]
- carbon_save: Estimated CO2 saved if recycled (for example: "50g CO2")
- tips: One short tip for eco-friendly disposal`),
		}, genai.RoleUser),
	}

	resp, err := h.client.Models.GenerateContent(
		context.Background(),
		h.model,
		contents,
		&genai.GenerateContentConfig{ResponseMIMEType: "application/json"},
	)
	if err != nil {
		return ScanResult{}, err
	}

	// Parse JSON response
	var result ScanResult
	resultText := strings.TrimSpace(resp.Text())
	if resultText == "" {
		return ScanResult{}, fmt.Errorf("gemini returned an empty response")
	}

	resultText = stripMarkdownFence(resultText)

	if err := json.Unmarshal([]byte(resultText), &result); err != nil {
		return ScanResult{}, fmt.Errorf("failed to parse Gemini JSON: %w", err)
	}

	return result, nil
}

func (h *ScanHandler) analyzeWithAzure(fileBytes []byte, contentType string) (ScanResult, error) {
	if !h.azure.enabled() {
		return ScanResult{}, fmt.Errorf("azure openai is not configured")
	}

	imageDataURL := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(fileBytes)

	requestBody := map[string]any{
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": `Analyze this item for Earth Day sustainability.
Return JSON with these exact keys:
- item
- material
- disposal (Recycle, Compost, Landfill, Hazardous, Reuse)
- carbon_save
- tips
Return only valid JSON.`,
					},
					{
						"type":      "image_url",
						"image_url": map[string]any{"url": imageDataURL},
					},
				},
			},
		},
		"temperature":     0.2,
		"max_tokens":      220,
		"response_format": map[string]any{"type": "json_object"},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return ScanResult{}, fmt.Errorf("marshal azure request: %w", err)
	}

	endpoint := normalizeAzureEndpoint(h.azure.Endpoint)
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", endpoint, h.azure.Deployment, h.azure.APIVersion)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return ScanResult{}, fmt.Errorf("create azure request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", h.azure.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ScanResult{}, fmt.Errorf("call azure openai: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ScanResult{}, fmt.Errorf("azure openai status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &decoded); err != nil {
		return ScanResult{}, fmt.Errorf("decode azure response: %w", err)
	}

	if len(decoded.Choices) == 0 {
		return ScanResult{}, fmt.Errorf("azure response has no choices")
	}

	contentText := extractAzureContent(decoded.Choices[0].Message.Content)
	if contentText == "" {
		return ScanResult{}, fmt.Errorf("azure response content is empty")
	}

	contentText = stripMarkdownFence(contentText)

	var result ScanResult
	if err := json.Unmarshal([]byte(contentText), &result); err != nil {
		return ScanResult{}, fmt.Errorf("parse azure JSON: %w", err)
	}

	return result, nil
}

func normalizeAzureEndpoint(raw string) string {
	trimmed := strings.TrimSpace(strings.TrimRight(raw, "/"))
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return trimmed
	}

	return parsed.Scheme + "://" + parsed.Host
}

func extractAzureContent(content any) string {
	switch v := content.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		for _, item := range v {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := part["text"].(string); ok {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func truncateError(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
