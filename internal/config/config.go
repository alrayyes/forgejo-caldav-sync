// Package config reads this tool's environment-variable configuration.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// errMissingRequiredEnv is wrapped with the actual comma-joined variable
// names rather than building a one-off dynamic error, so callers can match
// on it with errors.Is if they ever need to.
var errMissingRequiredEnv = errors.New("config: missing required environment variables")

// Config is the fully resolved, validated set of settings this tool needs
// to run.
type Config struct {
	ForgejoBaseURL       string
	ForgejoToken         string
	ForgejoWebhookSecret string

	CalDAVURL      string
	CalDAVUsername string
	CalDAVPassword string

	// Assignee restricts sync to issues assigned to this username. Empty
	// means every issue syncs, regardless of assignee.
	Assignee string

	Addr              string
	ReconcileInterval time.Duration
}

const (
	defaultAddr              = ":8080"
	defaultReconcileInterval = 15 * time.Minute
)

// Load reads Config from environment variables via getenv (os.Getenv in
// production; a fake map in tests).
func Load(getenv func(string) string) (Config, error) {
	var cfg Config
	var missing []string

	required := func(key string) string {
		v := getenv(key)
		if v == "" {
			missing = append(missing, key)
		}

		return v
	}

	cfg.ForgejoBaseURL = required("FORGEJO_BASE_URL")
	cfg.ForgejoToken = required("FORGEJO_TOKEN")
	cfg.ForgejoWebhookSecret = required("FORGEJO_WEBHOOK_SECRET")
	cfg.CalDAVURL = required("CALDAV_URL")
	cfg.CalDAVUsername = required("CALDAV_USERNAME")
	cfg.CalDAVPassword = required("CALDAV_PASSWORD")

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("%w: %s", errMissingRequiredEnv, strings.Join(missing, ", "))
	}

	cfg.Assignee = getenv("ASSIGNEE")

	cfg.Addr = defaultAddr
	if v := getenv("ADDR"); v != "" {
		cfg.Addr = v
	}

	cfg.ReconcileInterval = defaultReconcileInterval
	if v := getenv("RECONCILE_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("config: RECONCILE_INTERVAL: %w", err)
		}
		cfg.ReconcileInterval = d
	}

	return cfg, nil
}
