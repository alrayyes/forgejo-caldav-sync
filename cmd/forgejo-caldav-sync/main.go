// Command forgejo-caldav-sync serves a Forgejo issue webhook that upserts a
// matching CalDAV VTODO in real time, and periodically reconciles every
// issue on the instance against the calendar — the reconciliation pass is
// also what backfills older issues on first run, since it's the same code
// path either way.
//
// Every setting is an environment variable — see the README — so there is
// no flag parsing here. The one exception is a single positional
// "healthcheck" argument, invoked by the Dockerfile's HEALTHCHECK rather
// than a person, since the distroless base image has no shell or curl to
// run one out of.
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

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := runHealthcheck(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(context.Background(), logger); err != nil {
		logger.Error("forgejo-caldav-sync exiting", "error", err)
		os.Exit(1)
	}
}

func runHealthcheck() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	return checkHealth(cfg.Addr)
}

// checkHealth is what the "healthcheck" subcommand runs: a GET against this
// same process's own /healthz over loopback, since there's a live HTTP
// server to ask rather than a heartbeat file to read.
func checkHealth(addr string) error {
	resp, err := http.Get("http://localhost" + addr + "/healthz")
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: unexpected status %s", resp.Status)
	}
	return nil
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

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
