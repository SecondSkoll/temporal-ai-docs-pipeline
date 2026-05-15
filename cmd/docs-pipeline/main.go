package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
	handler "github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := controlplane.LoadConfig()

	if cfg.PostgresDSN == "" {
		logger.Error("POSTGRES_DSN environment variable is required")
		os.Exit(1)
	}

	// --- PostgreSQL ---
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}

	ledger := controlplane.NewLedger(pool)

	// --- Temporal client ---
	tc, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHostPort,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		logger.Error("failed to connect to temporal", "error", err)
		os.Exit(1)
	}
	defer tc.Close()

	// --- Temporal worker ---
	la := &controlplane.LedgerActivities{Ledger: ledger}
	w := worker.New(tc, controlplane.TaskQueueOrchestration, worker.Options{})
	w.RegisterWorkflow(controlplane.RunDocsWorkflow)
	w.RegisterActivity(controlplane.CheckSourceAccessActivity)
	w.RegisterActivity(la)

	go func() {
		if err := w.Run(worker.InterruptCh()); err != nil {
			logger.Error("temporal worker error", "error", err)
		}
	}()

	// --- HTTP server ---
	mux := http.NewServeMux()
	startHandler := &handler.StartHandler{
		Temporal: tc,
		Ledger:   ledger,
	}
	mux.HandleFunc("/v1/docs/runs", startHandler.HandleStart)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// --- Graceful shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server starting", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}
}
