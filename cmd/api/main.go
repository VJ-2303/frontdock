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

	"github.com/VJ-2303/frontdock/internal/config"
	"github.com/VJ-2303/frontdock/internal/database"
	"github.com/VJ-2303/frontdock/internal/httpx"
	"github.com/VJ-2303/frontdock/internal/users"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	godotenv.Load()
	cfg, err := config.Load(config.ServiceAPI)
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	usersStore := users.NewStore(pool)
	userHandler := users.NewHandler(usersStore, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth Routes
	mux.HandleFunc("POST /auth/register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", userHandler.Login)

	srv := &http.Server{
		Addr:              cfg.APIHTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("api listening", "addr", cfg.APIHTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "err", err)
		}
	}()

	<-stop
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
