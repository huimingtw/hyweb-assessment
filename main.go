// @title           HyWeb Assessment API
// @version         1.0
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huimingz/hyweb-assessment/config"
	"github.com/huimingz/hyweb-assessment/db"
	_ "github.com/huimingz/hyweb-assessment/docs"
	"github.com/huimingz/hyweb-assessment/handler"
	"github.com/huimingz/hyweb-assessment/router"
	"github.com/huimingz/hyweb-assessment/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	database, err := db.New(cfg.DSN, logger)
	if err != nil {
		logger.Error("db init failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	authSvc := service.NewAuthService(database, logger)
	userSvc := service.NewUserService(database, logger)
	weatherSvc := service.NewWeatherService(database, logger, cfg.WeatherAPIKey,
		&http.Client{Timeout: 10 * time.Second})

	go func() {
		weatherSvc.FetchAndStore(context.Background())
		for {
			time.Sleep(24 * time.Hour)
			weatherSvc.FetchAndStore(context.Background())
		}
	}()

	h := handler.NewHandler(authSvc, userSvc, weatherSvc, cfg, logger)
	r := router.New(h, cfg, logger)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("server stopped")
}
