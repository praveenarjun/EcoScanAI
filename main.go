package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"

	"ecoscan-ai/cache"
	"ecoscan-ai/config"
	"ecoscan-ai/handlers"
	"ecoscan-ai/middleware"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load local .env for development runs.
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file loaded from current directory", "error", err.Error())
	}

	slog.Info("🌍 Starting EcoScan AI...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err.Error())
		os.Exit(1)
	}

	slog.Info(
		"Active AI configuration",
		"provider", cfg.AIProvider,
		"gemini_model", cfg.ModelName,
		"azure_endpoint", cfg.AzureEndpoint,
		"azure_deployment", cfg.AzureDeployment,
	)

	var client *genai.Client
	if cfg.GeminiAPIKey != "" {
		client, err = genai.NewClient(context.Background(), &genai.ClientConfig{
			APIKey:  cfg.GeminiAPIKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			slog.Error("Failed to initialize Gemini", "error", err.Error())
			os.Exit(1)
		}
	}

	// Initialize cache
	appCache := cache.New(cfg.CacheTTL)

	// Initialize handlers
	scanHandler := handlers.NewScanHandler(
		client,
		appCache,
		cfg.ModelName,
		cfg.AIProvider,
		handlers.AzureConfig{
			Endpoint:   cfg.AzureEndpoint,
			APIKey:     cfg.AzureAPIKey,
			Deployment: cfg.AzureDeployment,
			APIVersion: cfg.AzureAPIVersion,
		},
	)

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Security Headers
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		if cfg.Environment != "production" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Rate Limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitPerSec)
	r.Use(rateLimiter.Middleware())

	// Routes
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/api/scan", scanHandler.Scan)

	// Health Check (Modern practice)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	// Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("Server running", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err.Error())
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err.Error())
	}

	slog.Info("Server exited cleanly", "provider", cfg.AIProvider, "model", cfg.ModelName)
}
