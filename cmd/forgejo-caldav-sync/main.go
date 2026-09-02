// Command forgejo-caldav-sync serves a Forgejo issue webhook that upserts a
// matching CalDAV VTODO in real time, and periodically reconciles every
// issue on the instance against the calendar — the reconciliation pass is
// also what backfills older issues on first run, since it's the same code
// path either way.
//
// Settings come from flags, environment variables and a config file, in
// that order of precedence — run `forgejo-caldav-sync --help` or see
// README.md for the full list. `init` writes a starter config file.
package main

import "os"

func main() {
	cmd := newRootCmd()
	cmd.CompletionOptions.DisableDefaultCmd = true
	if err := cmd.Execute(); err != nil {
		// cobra already printed the error to stderr (SilenceErrors is
		// false, the default) — this just sets the exit code.
		os.Exit(1)
	}
}
