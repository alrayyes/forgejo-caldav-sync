package main

import (
	"fmt"
	"os"

	"github.com/alrayyes/forgejo-caldav-sync/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// version is set via -ldflags "-X main.version=..." at build time; "dev"
// is what a local `go build`/`go run` gets instead.
var version = "dev"

// configFlags lists every setting Config carries, in flag/mapstructure-key
// order — this is the single source of truth addConfigFlags and
// bindConfigSources both read from, so a flag, its env var and its
// mapstructure key can't drift apart.
type configFlag struct {
	key   string // flag name == viper key == mapstructure tag
	env   string
	usage string
}

var configFlags = []configFlag{
	{"forgejo-base-url", "FORGEJO_BASE_URL", "Forgejo instance base URL"},
	{"forgejo-token", "FORGEJO_TOKEN", "Forgejo API token"},
	{"forgejo-webhook-secret", "FORGEJO_WEBHOOK_SECRET", "Forgejo webhook HMAC secret"},
	{"caldav-url", "CALDAV_URL", "CalDAV collection (task list) URL"},
	{"caldav-username", "CALDAV_USERNAME", "CalDAV basic auth username"},
	{"caldav-password", "CALDAV_PASSWORD", "CalDAV basic auth password"},
	{"assignee", "ASSIGNEE", "restrict sync to this assignee's issues (unset syncs every issue)"},
	{"addr", "ADDR", "address the HTTP server listens on"},
	{"reconcile-interval", "RECONCILE_INTERVAL", "how often the reconciliation pass runs"},
}

func newRootCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:           "forgejo-caldav-sync",
		Short:         "Sync Forgejo issues into a CalDAV calendar as VTODOs",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := resolveConfig(cmd, configPath)
			if err != nil {
				return err
			}

			return runServe(cmd.Context(), cfg)
		},
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default: XDG config dir)")
	addConfigFlags(cmd)

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newHealthcheckCmd(&configPath))

	return cmd
}

// addConfigFlags registers every Config setting as a persistent flag on
// cmd — inherited by every subcommand (healthcheck needs the same config
// resolution root's RunE does) — each defaulting to "" (or the zero
// duration) so viper, not pflag, owns the real default: a flag whose
// value is still its own zero value is exactly what tells viper "nothing
// was passed here, fall through to env, then file, then the real
// default."
func addConfigFlags(cmd *cobra.Command) {
	for _, f := range configFlags {
		if f.key == "reconcile-interval" {
			cmd.PersistentFlags().Duration(f.key, 0, f.usage)

			continue
		}
		cmd.PersistentFlags().String(f.key, "", f.usage)
	}
}

// resolveConfig layers flags over environment over a config file over
// built-in defaults — rules/cli.md's precedence order — and validates the
// result.
func resolveConfig(cmd *cobra.Command, configPath string) (config.Config, error) {
	v := viper.New()

	v.SetDefault("addr", config.DefaultAddr)
	v.SetDefault("reconcile-interval", config.DefaultReconcileInterval)

	for _, f := range configFlags {
		if err := v.BindPFlag(f.key, cmd.Flags().Lookup(f.key)); err != nil {
			return config.Config{}, fmt.Errorf("binding flag %s: %w", f.key, err)
		}
		if err := v.BindEnv(f.key, f.env); err != nil {
			return config.Config{}, fmt.Errorf("binding env %s: %w", f.env, err)
		}
	}

	foundFile, err := readConfigFile(v, configPath)
	if err != nil {
		return config.Config{}, err
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return config.Config{}, fmt.Errorf("decoding config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		if !foundFile && isTerminal(os.Stdin) {
			offerInit()
		}

		return config.Config{}, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}
