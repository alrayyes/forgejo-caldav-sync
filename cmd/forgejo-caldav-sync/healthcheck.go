package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// errUnexpectedStatus is wrapped with the actual status text rather than
// building a one-off dynamic error, so callers can match on it with
// errors.Is if they ever need to.
var errUnexpectedStatus = errors.New("unexpected status")

// newHealthcheckCmd builds the "healthcheck" subcommand: a GET against
// this same process's own /healthz over loopback, invoked by the
// Dockerfile's HEALTHCHECK rather than a person — hidden since the
// distroless base image has no shell or curl to run one out of, so this
// is what stands in for one.
func newHealthcheckCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:    "healthcheck",
		Short:  "Check that this process's own HTTP server is up",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := resolveConfig(cmd, *configPath)
			if err != nil {
				return fmt.Errorf("healthcheck: loading config: %w", err)
			}

			return checkHealth(cfg.Addr)
		},
	}
}

func checkHealth(addr string) error {
	resp, err := http.Get("http://localhost" + addr + "/healthz")
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: %w: %s", errUnexpectedStatus, resp.Status)
	}

	return nil
}
