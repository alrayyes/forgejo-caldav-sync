package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alrayyes/forgejo-caldav-sync/internal/api"
	"github.com/alrayyes/forgejo-caldav-sync/internal/caldav"
	"github.com/alrayyes/forgejo-caldav-sync/internal/config"
	"github.com/alrayyes/forgejo-caldav-sync/internal/forgejo"
	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
)

// runServe starts the webhook receiver and the reconciliation loop, and
// blocks until ctx is cancelled or the HTTP server fails.
func runServe(ctx context.Context, cfg config.Config) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	fg := forgejo.NewClient(cfg.ForgejoBaseURL, cfg.ForgejoToken)
	cal := caldav.NewClient(cfg.CalDAVURL, cfg.CalDAVUsername, cfg.CalDAVPassword)

	if err := cal.EnsureCollection(ctx); err != nil {
		return fmt.Errorf("ensuring caldav collection exists: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.NewMux(cal, cfg.ForgejoWebhookSecret, cfg.Assignee),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// context.Background() below, not ctx: ctx is already Done by the time
	// this goroutine runs, so deriving the shutdown timeout from it would
	// expire immediately instead of giving Shutdown its own 5 seconds.
	go func() { //nolint:gosec // see above
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("http shutdown", "error", err)
		}
	}()

	go reconcileLoop(ctx, logger, fg, cal, cfg.Assignee, cfg.ReconcileInterval)

	logger.Info("starting",
		"addr", cfg.Addr,
		"assignee", cfg.Assignee,
		"reconcile_interval", cfg.ReconcileInterval,
	)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

// reconcileLoop runs the reconciliation pass immediately — this is what
// backfills older issues on a cold start — and then on every tick of
// interval, until ctx is done.
func reconcileLoop(ctx context.Context, logger *slog.Logger, src sync.IssueSource, sink sync.CalendarSink, assignee string, interval time.Duration) {
	reconcileOnce(ctx, logger, src, sink, assignee)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcileOnce(ctx, logger, src, sink, assignee)
		}
	}
}

func reconcileOnce(ctx context.Context, logger *slog.Logger, src sync.IssueSource, sink sync.CalendarSink, assignee string) {
	n, err := sync.Reconcile(ctx, src, sink, assignee)
	if err != nil {
		logger.Error("reconcile failed, retrying next interval", "error", err)

		return
	}
	logger.Info("reconciled", "synced", n)
}
