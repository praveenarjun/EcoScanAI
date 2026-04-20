package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GeminiAPIKey    string
	Port            string
	Environment     string
	RateLimitPerSec int
	CacheTTL        time.Duration
	ModelName       string
	AIProvider      string
	AzureEndpoint   string
	AzureAPIKey     string
	AzureDeployment string
	AzureAPIVersion string
}

func Load() (*Config, error) {
	geminiKey := os.Getenv("GEMINI_API_KEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rps, err := strconv.Atoi(os.Getenv("RATE_LIMIT_PER_SECOND"))
	if err != nil || rps <= 0 {
		rps = 5
	}

	cacheTTLMinutes, err := strconv.Atoi(os.Getenv("CACHE_TTL_MINUTES"))
	if err != nil || cacheTTLMinutes <= 0 {
		cacheTTLMinutes = 60
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}

	aiProvider := os.Getenv("AI_PROVIDER")
	if aiProvider == "" {
		aiProvider = "gemini"
	}

	azureEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	azureAPIKey := os.Getenv("AZURE_OPENAI_API_KEY")
	azureDeployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")
	azureAPIVersion := os.Getenv("AZURE_OPENAI_API_VERSION")
	if azureAPIVersion == "" {
		azureAPIVersion = "2024-02-15-preview"
	}

	if aiProvider == "gemini" && geminiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when AI_PROVIDER=gemini")
	}

	if aiProvider == "azure" {
		if azureEndpoint == "" || azureAPIKey == "" || azureDeployment == "" {
			return nil, fmt.Errorf("AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_DEPLOYMENT are required when AI_PROVIDER=azure")
		}
	}

	if geminiKey == "" && (azureEndpoint == "" || azureAPIKey == "" || azureDeployment == "") {
		return nil, fmt.Errorf("configure either Gemini key or Azure OpenAI credentials")
	}

	return &Config{
		GeminiAPIKey:    geminiKey,
		Port:            port,
		Environment:     os.Getenv("ENVIRONMENT"),
		RateLimitPerSec: rps,
		CacheTTL:        time.Duration(cacheTTLMinutes) * time.Minute,
		ModelName:       modelName,
		AIProvider:      aiProvider,
		AzureEndpoint:   azureEndpoint,
		AzureAPIKey:     azureAPIKey,
		AzureDeployment: azureDeployment,
		AzureAPIVersion: azureAPIVersion,
	}, nil
}
