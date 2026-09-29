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

	"github.com/Simpleshaikh1/shipyard/internal/api"
	"github.com/Simpleshaikh1/shipyard/internal/config"
	"github.com/Simpleshaikh1/shipyard/internal/jobs"
)

func main() {
	cfg := config.Load()

	var level slog.Level
	_ = level.UnmarshalText([]byte(cfg.LogLevel))
	// JSON logs to stdout: Docker and Kubernetes collect stdout for you.
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.New(jobs.NewMemoryStore(), log).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ctx is cancelled when we get Ctrl+C (SIGINT) or SIGTERM.
	// SIGTERM is what `docker stop` and Kubernetes send before killing a container.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("api listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done() // block here until a shutdown signal arrives
	log.Info("shutting down")

	// Give in-flight requests up to 10s to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
	}
	log.Info("bye")
}
