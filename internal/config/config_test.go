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

	t.Run("succeeds", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, err)
	})

	t.Run("reads the required Forgejo variables", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "https://forge.example.com", cfg.ForgejoBaseURL)
		require.Equal(t, "forgejo-token", cfg.ForgejoToken)
		require.Equal(t, "webhook-secret", cfg.ForgejoWebhookSecret)
	})

	t.Run("reads the required CalDAV variables", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "https://dav.example.com/calendars/alice/forgejo/", cfg.CalDAVURL)
		require.Equal(t, "alice", cfg.CalDAVUsername)
		require.Equal(t, "caldav-password", cfg.CalDAVPassword)
	})

	t.Run("assignee defaults to unset", func(t *testing.T) {
		t.Parallel()
		require.Empty(t, cfg.Assignee)
	})

	t.Run("addr defaults to :8080", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, ":8080", cfg.Addr)
	})

	t.Run("reconcile interval defaults to 15 minutes", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 15*time.Minute, cfg.ReconcileInterval)
	})
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

	t.Run("succeeds", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, err)
	})

	t.Run("assignee is overridden", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "bob", cfg.Assignee)
	})

	t.Run("addr is overridden", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, ":9090", cfg.Addr)
	})

	t.Run("reconcile interval is overridden", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 5*time.Minute, cfg.ReconcileInterval)
	})
}

func TestLoadRejectsAnInvalidReconcileInterval(t *testing.T) {
	t.Parallel()

	_, err := config.Load(fakeEnv(map[string]string{"RECONCILE_INTERVAL": "not-a-duration"}))

	require.ErrorContains(t, err, "RECONCILE_INTERVAL")
}
