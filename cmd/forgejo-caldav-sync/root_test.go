package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// setXDGConfigHome points xdg.ConfigFile at dir for the duration of the
// test. adrg/xdg reads $XDG_CONFIG_HOME once at package init, so a plain
// t.Setenv has no effect on it — xdg.Reload() forces a fresh read, and the
// cleanup reloads again so later tests see the real environment.
func setXDGConfigHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", dir)
	xdg.Reload()
	t.Cleanup(xdg.Reload)
}

// newTestCmd returns a bare *cobra.Command carrying the same persistent
// config flags the real root command registers, for exercising
// resolveConfig without going through cmd.Execute().
func newTestCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	addConfigFlags(cmd)
	require.NoError(t, cmd.ParseFlags(nil))

	return cmd
}

func requiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("FORGEJO_BASE_URL", "https://forge.example.com")
	t.Setenv("FORGEJO_TOKEN", "forgejo-token")
	t.Setenv("FORGEJO_WEBHOOK_SECRET", "webhook-secret")
	t.Setenv("CALDAV_URL", "https://dav.example.com/calendars/alice/forgejo/")
	t.Setenv("CALDAV_USERNAME", "alice")
	t.Setenv("CALDAV_PASSWORD", "caldav-password")
}

// resolveConfig reads real environment variables (t.Setenv) and can read
// a real file from disk, so these can't run in parallel with each other
// without racing on shared process-wide state.

func TestResolveConfigReadsEnvironmentVariables(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())
	requiredEnv(t)

	cfg, err := resolveConfig(newTestCmd(t), "")

	require.NoError(t, err)
	require.Equal(t, "https://forge.example.com", cfg.ForgejoBaseURL)
}

func TestResolveConfigFlagBeatsEnv(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())
	requiredEnv(t)
	t.Setenv("ADDR", ":7777")

	cmd := newTestCmd(t)
	require.NoError(t, cmd.Flags().Set("addr", ":6666"))

	cfg, err := resolveConfig(cmd, "")

	require.NoError(t, err)
	require.Equal(t, ":6666", cfg.Addr)
}

func TestResolveConfigEnvBeatsFile(t *testing.T) {
	dir := t.TempDir()
	setXDGConfigHome(t, dir)
	requiredEnv(t)
	t.Setenv("ADDR", ":7777")

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "forgejo-caldav-sync"), 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "forgejo-caldav-sync", "config.yaml"),
		[]byte("addr: \":8888\"\n"), 0o600,
	))

	cfg, err := resolveConfig(newTestCmd(t), "")

	require.NoError(t, err)
	require.Equal(t, ":7777", cfg.Addr)
}

func TestResolveConfigFileBeatsDefault(t *testing.T) {
	dir := t.TempDir()
	setXDGConfigHome(t, dir)
	requiredEnv(t)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "forgejo-caldav-sync"), 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "forgejo-caldav-sync", "config.yaml"),
		[]byte("reconcile-interval: \"5m\"\n"), 0o600,
	))

	cfg, err := resolveConfig(newTestCmd(t), "")

	require.NoError(t, err)
	require.Equal(t, 5*time.Minute, cfg.ReconcileInterval)
}

func TestResolveConfigDefaultsWithNothingSet(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())
	requiredEnv(t)

	cfg, err := resolveConfig(newTestCmd(t), "")

	require.NoError(t, err)
	require.Equal(t, ":8080", cfg.Addr)
}

func TestResolveConfigMissingRequiredSettingErrors(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())

	_, err := resolveConfig(newTestCmd(t), "")

	require.Error(t, err)
}

func TestResolveConfigExplicitMissingFileErrors(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())
	requiredEnv(t)

	_, err := resolveConfig(newTestCmd(t), filepath.Join(t.TempDir(), "does-not-exist.yaml"))

	require.Error(t, err)
}
