// Package config is the fully resolved, validated set of settings this
// tool needs to run. Resolving it from flags, the environment and a config
// file is cmd/forgejo-caldav-sync's job — this package only knows the
// shape of the result and how to validate it.
package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// ErrMissingRequired is wrapped with the actual comma-joined setting names
// rather than building a one-off dynamic error, so callers can match on it
// with errors.Is if they ever need to.
var ErrMissingRequired = errors.New("config: missing required settings")

// Config is the fully resolved set of settings this tool needs to run.
// Field tags are the mapstructure keys viper.Unmarshal reads them from —
// also the flag names and config-file keys, so all three stay in sync by
// construction.
type Config struct {
	ForgejoBaseURL       string `mapstructure:"forgejo-base-url"`
	ForgejoToken         string `mapstructure:"forgejo-token"`
	ForgejoWebhookSecret string `mapstructure:"forgejo-webhook-secret"`

	CalDAVURL      string `mapstructure:"caldav-url"`
	CalDAVUsername string `mapstructure:"caldav-username"`
	CalDAVPassword string `mapstructure:"caldav-password"`

	// Assignee restricts sync to issues assigned to this username. Empty
	// means every issue syncs, regardless of assignee.
	Assignee string `mapstructure:"assignee"`

	Addr              string        `mapstructure:"addr"`
	ReconcileInterval time.Duration `mapstructure:"reconcile-interval"`
}

// Defaults are what a config file, flag or environment variable overrides,
// and what `init` writes into a starter config file.
const (
	DefaultAddr              = ":8080"
	DefaultReconcileInterval = 15 * time.Minute
)

// Validate reports every required setting that's still empty once flags,
// environment and config file have all been applied. Called once right
// after the config is resolved, so a bad or missing value fails at
// startup instead of wherever it's first read.
func (c Config) Validate() error {
	required := map[string]string{
		"forgejo-base-url":       c.ForgejoBaseURL,
		"forgejo-token":          c.ForgejoToken,
		"forgejo-webhook-secret": c.ForgejoWebhookSecret,
		"caldav-url":             c.CalDAVURL,
		"caldav-username":        c.CalDAVUsername,
		"caldav-password":        c.CalDAVPassword,
	}

	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	slices.Sort(missing)

	return fmt.Errorf("%w: %s", ErrMissingRequired, strings.Join(missing, ", "))
}
