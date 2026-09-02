package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunInitWritesAConfigFile(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())

	var out bytes.Buffer
	err := runInit(&out, false)

	require.NoError(t, err)
}

func TestRunInitWrittenFileParsesAsYAML(t *testing.T) {
	dir := t.TempDir()
	setXDGConfigHome(t, dir)

	require.NoError(t, runInit(&bytes.Buffer{}, false))

	path := filepath.Join(dir, "forgejo-caldav-sync", "config.yaml")
	data, err := os.ReadFile(path) //nolint:gosec // reading a path this test wrote, inside its own t.TempDir()
	require.NoError(t, err)
	require.Contains(t, string(data), "addr:")
}

func TestRunInitRefusesToOverwriteWithoutForce(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())
	require.NoError(t, runInit(&bytes.Buffer{}, false))

	err := runInit(&bytes.Buffer{}, false)

	require.Error(t, err)
}

func TestRunInitForceOverwrites(t *testing.T) {
	setXDGConfigHome(t, t.TempDir())
	require.NoError(t, runInit(&bytes.Buffer{}, false))

	err := runInit(&bytes.Buffer{}, true)

	require.NoError(t, err)
}
