package config_test

import (
	"testing"

	"github.com/alrayyes/forgejo-caldav-sync/internal/config"
	"github.com/stretchr/testify/require"
)

func validConfig() config.Config {
	return config.Config{
		ForgejoBaseURL:       "https://forge.example.com",
		ForgejoToken:         "forgejo-token",
		ForgejoWebhookSecret: "webhook-secret",
		CalDAVURL:            "https://dav.example.com/calendars/alice/forgejo/",
		CalDAVUsername:       "alice",
		CalDAVPassword:       "caldav-password",
		Addr:                 config.DefaultAddr,
		ReconcileInterval:    config.DefaultReconcileInterval,
	}
}

func TestValidateAcceptsEveryRequiredSetting(t *testing.T) {
	t.Parallel()

	require.NoError(t, validConfig().Validate())
}

func TestValidateReportsEveryMissingRequiredSetting(t *testing.T) {
	t.Parallel()

	err := config.Config{}.Validate()

	require.ErrorContains(t, err, "forgejo-base-url")
	require.ErrorContains(t, err, "caldav-url")
}

func TestValidateReportsOneMissingSetting(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.ForgejoToken = ""

	require.ErrorContains(t, cfg.Validate(), "forgejo-token")
}

func TestValidateIsIndifferentToOptionalSettings(t *testing.T) {
	t.Parallel()

	cfg := validConfig()
	cfg.Assignee = ""

	require.NoError(t, cfg.Validate())
}
