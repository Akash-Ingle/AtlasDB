package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	aiPkg "github.com/atlasdb/atlasdb/internal/ai"
	"github.com/atlasdb/atlasdb/internal/alerts"
	"github.com/atlasdb/atlasdb/internal/analytics"
	"github.com/atlasdb/atlasdb/internal/api"
	"github.com/atlasdb/atlasdb/internal/api/handlers"
	"github.com/atlasdb/atlasdb/internal/auth"
	"github.com/atlasdb/atlasdb/internal/config"
	"github.com/atlasdb/atlasdb/internal/ingest"
	"github.com/atlasdb/atlasdb/internal/queue"
	"github.com/atlasdb/atlasdb/internal/search"
	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	redisstore "github.com/atlasdb/atlasdb/internal/storage/redis"
	"github.com/atlasdb/atlasdb/internal/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := telemetry.InitLogger(cfg.Log.Level, cfg.Log.Format, "api-server")
	logger.Info().Msg("Starting AtlasDB API server")

	// OpenTelemetry
	shutdownTracer, err := telemetry.InitTracer(ctx, telemetry.TracerConfig{
		ServiceName:    "atlas-api-server",
		ServiceVersion: "1.0.0",
		OTLPEndpoint:   cfg.OTel.OTLPEndpoint,
		Enabled:        cfg.OTel.Enabled,
	})
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to init tracer, continuing without tracing")
	} else {
		defer shutdownTracer(ctx)
	}

	metrics := telemetry.NewMetrics("atlas")

	// PostgreSQL
	pgPool, err := postgres.NewPool(ctx, cfg.Postgres, logger)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pgPool.Close()

	if err := postgres.RunMigrations(ctx, pgPool, logger); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	// Redis
	redisClient, err := redisstore.NewClient(ctx, cfg.Redis, logger)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer redisClient.Close()

	// Stores
	eventStore := postgres.NewEventStore(pgPool, logger)
	userStore := postgres.NewUserStore(pgPool, logger)
	apiKeyStore := auth.NewAPIKeyStore(pgPool)
	alertStore := alerts.NewStore(pgPool, logger)

	// Services
	producer := queue.NewProducer(redisClient, cfg.Ingest.NumPartitions, logger)
	ingester := ingest.NewIngester(producer, cfg.Ingest, metrics, logger)
	searchEngine := search.NewEngine(pgPool, logger)
	analyticsEngine := analytics.NewEngine(pgPool, logger)
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)

	// WebSocket
	broadcaster := handlers.NewEventBroadcaster(logger)

	// Handlers
	eventHandler := handlers.NewEventHandler(ingester, eventStore, logger)
	authHandler := handlers.NewAuthHandler(userStore, jwtManager, logger)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyStore, logger)
	searchHandler := handlers.NewSearchHandler(searchEngine, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsEngine, logger)
	alertHandler := handlers.NewAlertHandler(alertStore, logger)
	dlqManager := queue.NewDLQManager(redisClient, logger)
	adminHandler := handlers.NewAdminHandler(dlqManager, cfg.Ingest.NumPartitions, logger)
	systemHandler := handlers.NewSystemHandler(pgPool, redisClient)
	wsHandler := handlers.NewWebSocketHandler(broadcaster, logger)

	// AI (optional — only if API key configured)
	var aiHandler *handlers.AIHandler
	if cfg.AI.Enabled {
		var llmProvider aiPkg.LLMProvider
		switch cfg.AI.Provider {
		case "anthropic":
			llmProvider = aiPkg.NewAnthropicProvider(aiPkg.AnthropicConfig{
				APIKey: cfg.AI.AnthropicKey,
				Model:  cfg.AI.AnthropicModel,
			}, logger)
		default:
			llmProvider = aiPkg.NewOpenAIProvider(aiPkg.OpenAIConfig{
				APIKey:     cfg.AI.OpenAIKey,
				BaseURL:    cfg.AI.OpenAIBaseURL,
				Model:      cfg.AI.OpenAIModel,
				EmbedModel: cfg.AI.EmbedModel,
			}, logger)
		}

		queryEngine := aiPkg.NewQueryEngine(llmProvider, logger)
		toolRegistry := aiPkg.NewToolRegistry(eventStore, analyticsEngine, alertStore, logger)
		investigator := aiPkg.NewInvestigator(llmProvider, toolRegistry, logger)
		anomalyDet := aiPkg.NewAnomalyDetector(pgPool, llmProvider, logger)

		aiHandler = handlers.NewAIHandler(queryEngine, investigator, anomalyDet, logger)
		logger.Info().Str("provider", cfg.AI.Provider).Msg("AI layer enabled")
	} else {
		logger.Info().Msg("AI layer disabled (set AI_ENABLED=true to enable)")
	}

	router := api.NewRouter(api.RouterDeps{
		EventHandler:     eventHandler,
		AuthHandler:      authHandler,
		APIKeyHandler:    apiKeyHandler,
		SearchHandler:    searchHandler,
		AnalyticsHandler: analyticsHandler,
		AlertHandler:     alertHandler,
		AIHandler:        aiHandler,
		AdminHandler:     adminHandler,
		SystemHandler:    systemHandler,
		WebSocketHandler: wsHandler,
		JWTManager:       jwtManager,
		Metrics:          metrics,
		Logger:           logger,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info().Str("addr", addr).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info().Msg("Shutdown signal received")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	logger.Info().Msg("Server stopped gracefully")
	return nil
}
