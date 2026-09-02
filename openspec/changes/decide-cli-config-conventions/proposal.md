## Why

`rules/go.md`'s Command line/Configuration sections and
`rules/cli.md` (which loads on any `cmd/**/main.go`) call for
spf13/cobra + spf13/viper, a config-file layer, a `init` command, and
XDG-spec directories. This tool is a 12-factor-style Docker daemon —
every setting is an environment variable, no flags, no config file,
no `init` command — a deliberate choice already documented in
`main.go`'s own doc comment. `rules/cli.md` reads as written with an
interactive CLI in mind, not obviously a long-running webhook
receiver/reconciliation daemon meant to run one way, in a container.
Found during a repo-wide audit against `~/.config/claude/CLAUDE.md`
and its `rules/*.md` — filed as a decision rather than changed blind.

**Not yet decided.** See
[issue #30](https://github.com/alrayyes/forgejo-caldav-sync/issues/30)
for the acceptance criteria this proposal has to satisfy either way.

## What Changes

Either:

- Adopt cobra + viper: flags/env/config-file precedence follows
  `rules/cli.md` (flags > env > file > defaults), a
  `forgejo-caldav-sync init` command writes a starter config, and the
  existing env-var-only behaviour keeps working for anyone not
  touching the new file/flag layer.

or:

- Keep the current env-var-only design, with `main.go`'s existing doc
  comment standing as the documented reason, so a future audit
  doesn't re-flag it.

## Capabilities

### New Capabilities

- `configuration` (only if adopted): a config-file layer and an
  `init` command, layered under the existing environment-variable
  configuration per `rules/cli.md`'s precedence order.

### Modified Capabilities

(none — `issue-sync`'s behavior is unaffected either way; this is
about how configuration reaches it, not what it does with it)

## Impact

`cmd/forgejo-caldav-sync/main.go`, `internal/config`, potentially a
new `internal/cli` (cobra command tree) and a config-file schema.
Whether this needs a spec delta for a new `configuration` capability
is itself part of the decision — left open here rather than
prejudged.
