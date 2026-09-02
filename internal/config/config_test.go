package config_test

import (
	"maps"
	"testing"
	"time"

	"github.com/alrayyes/forgejo-caldav-sync/internal/config"
	"github.com/stretchr/testify/require"
)

func fakeEnv(overrides map[string]string) func(string) string {
	base := map[string]string{
		"FORGEJO_BASE_URL":       "https://forge.example.com",
		"FORGEJO_TOKEN":          "forgejo-token",
		"FORGEJO_WEBHOOK_SECRET": "webhook-secret",
		"CALDAV_URL":             "https://dav.example.com/calendars/alice/forgejo/",
		"CALDAV_USERNAME":        "alice",
		"CALDAV_PASSWORD":        "caldav-password",
	}
	maps.Copy(base, overrides)

	return func(key string) string { return base[key] }
}

func TestLoadWithOnlyRequiredVariables(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(fakeEnv(nil))

	require.NoError(t, err)
	require.Equal(t, "https://forge.example.com", cfg.ForgejoBaseURL)
	require.Equal(t, "forgejo-token", cfg.ForgejoToken)
	require.Equal(t, "webhook-secret", cfg.ForgejoWebhookSecret)
	require.Equal(t, "https://dav.example.com/calendars/alice/forgejo/", cfg.CalDAVURL)
	require.Equal(t, "alice", cfg.CalDAVUsername)
	require.Equal(t, "caldav-password", cfg.CalDAVPassword)
	require.Empty(t, cfg.Assignee)
	require.Equal(t, ":8080", cfg.Addr)
	require.Equal(t, 15*time.Minute, cfg.ReconcileInterval)
}

func TestLoadReportsEveryMissingRequiredVariable(t *testing.T) {
	t.Parallel()

	_, err := config.Load(func(string) string { return "" })

	require.ErrorContains(t, err, "FORGEJO_BASE_URL")
	require.ErrorContains(t, err, "CALDAV_URL")
}

func TestLoadHonorsOptionalOverrides(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(fakeEnv(map[string]string{
		"ASSIGNEE":           "bob",
		"ADDR":               ":9090",
		"RECONCILE_INTERVAL": "5m",
	}))

	require.NoError(t, err)
	require.Equal(t, "bob", cfg.Assignee)
	require.Equal(t, ":9090", cfg.Addr)
	require.Equal(t, 5*time.Minute, cfg.ReconcileInterval)
}

func TestLoadRejectsAnInvalidReconcileInterval(t *testing.T) {
	t.Parallel()

	_, err := config.Load(fakeEnv(map[string]string{"RECONCILE_INTERVAL": "not-a-duration"}))

	require.ErrorContains(t, err, "RECONCILE_INTERVAL")
}
