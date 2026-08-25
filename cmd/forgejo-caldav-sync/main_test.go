package main

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckHealthAgainstAHealthyServer(t *testing.T) {
	addr := startHealthzServer(t, http.StatusOK)

	require.NoError(t, checkHealth(addr))
}

func TestCheckHealthReportsAnUnhealthyServer(t *testing.T) {
	addr := startHealthzServer(t, http.StatusServiceUnavailable)

	require.Error(t, checkHealth(addr))
}

func TestCheckHealthReportsAnUnreachableServer(t *testing.T) {
	require.Error(t, checkHealth(":1"))
}

// startHealthzServer runs a real HTTP server on loopback answering /healthz
// with status, and returns the ":port" form checkHealth expects.
func startHealthzServer(t *testing.T, status int) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	go func() { _ = http.Serve(ln, mux) }() //nolint:errcheck,gosec // server lifetime is the test's

	return fmt.Sprintf(":%d", ln.Addr().(*net.TCPAddr).Port)
}
