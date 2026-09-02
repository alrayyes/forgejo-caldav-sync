## Context

See proposal.md - Why. The tool needs to keep working exactly as
deployed today (Docker + env vars) while gaining a config-file/flag
layer for anyone running the binary directly.

## Goals / Non-Goals

**Goals**: zero behavior change for existing env-var-only deployments;
`spf13/viper`'s own precedence order does the layering, not hand-rolled
merge logic; keep `internal/config` framework-free.

**Non-goals**: a config format migration tool, config file encryption
or secret management beyond what env vars already had (plaintext),
multiple simultaneous config files.

## Decisions

- **`internal/config` stays framework-free.** It only defines `Config`
  and `Validate()` — no cobra or viper import. The actual flag
  definitions, env bindings and file resolution live in `cmd/`, since
  that layering is inherently tied to the command-line surface
  (`*pflag.FlagSet`, `*cobra.Command`) and `internal/config` shouldn't
  need to know cobra exists to be unit tested. Alternative considered:
  push the whole viper setup into `internal/config` for "one place to
  look" — rejected because it would make a hexagonal domain-support
  package depend on a CLI framework for no benefit.
- **A `configFlag` table (name/env/usage) is the single source of
  truth** `addConfigFlags` and `resolveConfig` both read from, so a
  flag, its env var and its mapstructure key can't drift apart
  independently. Alternative considered: define flags and bindings
  separately (as most cobra+viper examples do) — rejected after
  confirming a copy-paste flag/env mismatch is exactly the kind of bug
  that's invisible until someone hits it in production.
- **Flags default to their zero value, not the real default.** viper's
  `SetDefault` owns the real default (`:8080`,
  `15m`); a pflag default would make viper unable to tell "the user
  passed `--addr :8080`" apart from "nothing was passed," breaking the
  env/file layers underneath it. This is why every flag is
  registered with `""` (or zero `time.Duration`) rather than the
  actual default value.
- **`golang.org/x/term.IsTerminal`, not a heavier TTY-detection
  library**, for the first-run gate — it's already a transitive
  dependency of the toolchain's own use of terminal detection
  elsewhere in the Go ecosystem, single-purpose, and exactly what
  `rules/cli.md` names.
- **`adrg/xdg` for the config path, with `xdg.Reload()` used only in
  tests.** The package reads `$XDG_CONFIG_HOME` once at process init,
  which is correct and irrelevant for a real short-lived CLI
  invocation, but breaks a test's `t.Setenv` unless it forces a
  re-read. Confirmed live rather than assumed: the package exports
  `Reload()` for exactly this.
- **`healthcheck` becomes a hidden cobra subcommand**, not the
  `os.Args[1] == "healthcheck"` check it was — the Dockerfile's
  invocation (`/forgejo-caldav-sync healthcheck`) needed no change,
  confirmed against a real `docker build` + container run with Docker's
  own `HEALTHCHECK` reporting `healthy`.

## Risks / Trade-offs

- [A config file holds secrets in plaintext, same as `.env` already
  does] → Accepted; no worse than the status quo, and `init` writes
  the file `0o600`.
- [Two ways to configure the same tool is more surface to document] →
  Mitigated by keeping env vars as the primary documented path in
  README (Docker deployment) and the config file as the secondary path
  for a directly-run binary.

## Migration Plan

No migration needed — every existing `.env`/Docker deployment keeps
working unchanged, since environment variables are still read via the
same names at the same precedence they always effectively had (now
formalized as "below flags, above the config file").
