package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/config"
	"github.com/JuhaoChen666/RentNestHub/backend/internal/httpapi"
	"github.com/JuhaoChen666/RentNestHub/backend/internal/repository/mysqlrepo"
	"github.com/JuhaoChen666/RentNestHub/backend/internal/service"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	repository, err := mysqlrepo.New(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer repository.Close()

	recommender := service.NewRecommenderWithProvider(
		repository,
		service.NewRecommendationProvider(service.AIProviderConfig{
			URL:    cfg.AIAPIURL,
			APIKey: cfg.AIAPIKey,
			Model:  cfg.AIModel,
		}),
	)
	handler := httpapi.New(httpapi.Dependencies{
		Repository:    repository,
		Recommender:   recommender,
		UploadDir:     cfg.UploadDir,
		PublicBaseURL: cfg.PublicBaseURL,
		Logger:        logger,
	})

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve api", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown api", "error", err)
	}
}
