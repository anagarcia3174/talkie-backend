package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anagarcia3174/talkie-backend/internal/auth"
	"github.com/anagarcia3174/talkie-backend/internal/blocks"
	"github.com/anagarcia3174/talkie-backend/internal/comments"
	"github.com/anagarcia3174/talkie-backend/internal/config"
	"github.com/anagarcia3174/talkie-backend/internal/db"
	dbgen "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/anagarcia3174/talkie-backend/internal/httpapi"
	"github.com/anagarcia3174/talkie-backend/internal/logging"
	"github.com/anagarcia3174/talkie-backend/internal/reports"
	"github.com/anagarcia3174/talkie-backend/internal/tmdb"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logger := logging.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("error connecting to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := dbgen.New(pool)
	verifier, err := auth.NewVerifier(ctx, cfg.SupabaseJWKSURL)
	if err != nil {
		logger.Error("error setting up jwt verifier", "error", err)
		os.Exit(1)
	}
	adminClient := auth.NewAdminClient(cfg.SupabaseURL, cfg.SupabaseSecretKey)
	appleClient, err := auth.NewAppleClient(cfg.AppleTeamID, cfg.AppleKeyID, cfg.AppleClientID, cfg.ApplePrivateKey)
	if err != nil {
		logger.Error("error setting up apple client", "error", err)
		os.Exit(1)
	}
	authSvc := auth.NewService(queries, adminClient, appleClient, logger)
	authHandler := httpapi.NewAuthHandler(authSvc, logger)
	tmdbClient := tmdb.NewClient(cfg.TMDBAPIKey)
	mediaHandler := httpapi.NewMediaHandler(tmdbClient, logger)
	commentSvc := comments.NewService(queries)
	commentHandler := httpapi.NewCommentHandler(commentSvc, logger)
	reportSvc := reports.NewService(queries)
	reportHandler := httpapi.NewReportHandler(reportSvc, logger)
	blockSvc := blocks.NewService(queries)
	blockHandler := httpapi.NewBlockHandler(blockSvc, logger)
	router := httpapi.NewRouter(httpapi.RouterDeps{
		Logger:         logger,
		AuthHandler:    authHandler,
		MediaHandler:   mediaHandler,
		CommentHandler: commentHandler,
		ReportHandler:  reportHandler,
		BlockHandler:   blockHandler,
		Verifier:       verifier,
		BanChecker:     queries,
		TermsChecker:   queries,
		AllowedOrigin:  cfg.AllowedOrigin,
		ClientIPHeader: cfg.ClientIPHeader,
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	logger.Info("starting server", "port", cfg.Port)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		logger.Info("server shut down cleanly")
	}
}
