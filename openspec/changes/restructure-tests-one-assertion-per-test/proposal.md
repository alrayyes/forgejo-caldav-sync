## Why

`rules/go-test.md`: "One assertion per test, with `t.Run` subtests
for the scenarios around one thing. A test asserting five things
reports the first failure and hides the other four." Most of the 38
tests across `internal/sync`, `internal/api`, `internal/caldav`,
`internal/forgejo`, `internal/config` and `cmd/forgejo-caldav-sync`
assert several fields or behaviours per test function today. Found
during a repo-wide audit against `~/.config/claude/CLAUDE.md` and its
`rules/*.md` — filed as its own change rather than fixed inline,
since restructuring 38 tests across 6 files is a real rewrite with
its own risk of losing coverage or intent along the way.

See [issue #31](https://github.com/alrayyes/forgejo-caldav-sync/issues/31)
for the acceptance criteria.

## What Changes

- Tests that assert multiple unrelated fields or behaviours become a
  `t.Run` table or a set of sibling tests, each asserting one thing
  and named for what it checks.
- `go test -race ./...` stays green and coverage doesn't drop below
  what it is today.
- A genuinely single-scenario mapping test may keep several assertions
  when they're one coherent claim, not several unrelated ones bundled
  for convenience — judged case by case, not applied mechanically to
  every test.

## Capabilities

### New Capabilities

(none — this changes how behavior is verified, not the behavior
itself)

### Modified Capabilities

(none)

This is test-structure only, no product behavior changes:
`skip_specs: true` is set in `.openspec.yaml`.

## Impact

`internal/sync/sync_test.go`, `internal/api/api_test.go`,
`internal/caldav/caldav_test.go`, `internal/forgejo/forgejo_test.go`,
`internal/config/config_test.go`,
`cmd/forgejo-caldav-sync/main_test.go`.
