package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alrayyes/forgejo-caldav-sync/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// errConfigExists is wrapped with the actual path rather than building a
// one-off dynamic error, so callers can match on it with errors.Is if
// they ever need to.
var errConfigExists = errors.New("config file already exists")

// starterConfig is what `init` writes: every setting this tool reads,
// commented where a real value has to be filled in, populated with the
// defaults it would otherwise fall back to everywhere else.
var starterConfig = `# forgejo-caldav-sync config file.
# Flags and environment variables override whatever's set here — see
# README.md for the full list and what each setting means.

# forgejo-base-url: ""
# forgejo-token: ""
# forgejo-webhook-secret: ""

# caldav-url: ""
# caldav-username: ""
# caldav-password: ""

# assignee: ""       # unset: sync every issue, regardless of assignee
addr: "` + config.DefaultAddr + `"
reconcile-interval: "` + config.DefaultReconcileInterval.String() + `"
`

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write a starter config file",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInit(os.Stdout, force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing config file")

	return cmd
}

func runInit(out io.Writer, force bool) error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %s (use --force to overwrite)", errConfigExists, path)
		}
	}

	if err := os.WriteFile(path, []byte(starterConfig), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	if _, err := fmt.Fprintf(out, "Wrote %s — fill in the commented-out settings and re-run.\n", path); err != nil {
		return fmt.Errorf("writing init output: %w", err)
	}

	return nil
}

// isTerminal reports whether f is a real, interactive terminal. CI and any
// piped or scripted invocation has no TTY, and a prompt that blocks on
// input a script will never send hangs the job instead of failing it fast.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// offerInit is the first-run nudge: a genuinely unconfigured run, with a
// TTY to ask on, is exactly the moment someone needs pointing at `init`
// rather than left to find it in --help. It only ever runs once — after a
// config file exists, cfg.Validate() either passes or reports missing
// values from a file that's already there, not "you have no config at
// all."
func offerInit() {
	fmt.Fprint(os.Stderr, "No configuration found. Write a starter config file now? [y/N] ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	if answer := strings.ToLower(strings.TrimSpace(scanner.Text())); answer != "y" && answer != "yes" {
		return
	}

	if err := runInit(os.Stderr, false); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
