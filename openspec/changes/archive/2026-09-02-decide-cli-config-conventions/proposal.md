## Why

`rules/go.md`'s Command line/Configuration sections and `rules/cli.md`
(which loads on any `cmd/**/main.go`) call for spf13/cobra +
spf13/viper, a config-file layer, an `init` command, and XDG-spec
directories. This tool was a 12-factor-style Docker daemon — every
setting an environment variable, no flags, no config file, no `init`
command — a deliberate choice, documented in `main.go`'s own doc
comment. Found during a repo-wide audit against
`~/.config/claude/CLAUDE.md` and its `rules/*.md`, filed as a decision
rather than changed blind ([issue #30](https://github.com/alrayyes/forgejo-caldav-sync/issues/30)).

**Decided: adopt cobra + viper.** Environment variables stay fully
supported — nothing about the Docker-deployment path changes — the
config-file and flag layers are additive for anyone running the binary
directly.

## What Changes

- flags/env/config-file precedence follows `rules/cli.md` (flags > env
  > file > defaults), via `spf13/viper`'s own native precedence.
- `forgejo-caldav-sync init` writes a starter config file at
  `$XDG_CONFIG_HOME/forgejo-caldav-sync/config.yaml` (or `--config
<path>`).
- A run with no config file and no relevant environment variable set,
  at a real terminal, offers to write one — `golang.org/x/term` gates
  the prompt so a non-interactive run (CI, Docker) never blocks on
  stdin.
- `healthcheck` becomes a proper (hidden) cobra subcommand instead of
  a hand-rolled `os.Args[1]` check.
- `internal/config.Config` gets a `Validate() error` method, called
  once after the config is fully resolved.

## Capabilities

### New Capabilities

- `configuration`: the config-file/flag/env layering, the `init`
  command, and the first-run prompt.

### Modified Capabilities

(none — `issue-sync`'s behavior is unaffected; this changes how
configuration reaches it, not what it does with it)

## Impact

`cmd/forgejo-caldav-sync/main.go` (now a thin entrypoint),
`cmd/forgejo-caldav-sync/root.go` (new: cobra root command, flag
definitions, `resolveConfig`), `cmd/forgejo-caldav-sync/serve.go` (new:
the server + reconciliation loop, moved out of `main.go`),
`cmd/forgejo-caldav-sync/init.go` (new), `cmd/forgejo-caldav-sync/healthcheck.go`
(new), `internal/config` (now just the `Config` struct and
`Validate()` — the wiring lives in `cmd/` since it's inherently tied
to the command-line surface). New dependencies: `spf13/cobra`,
`spf13/viper`, `adrg/xdg`, `golang.org/x/term`.
