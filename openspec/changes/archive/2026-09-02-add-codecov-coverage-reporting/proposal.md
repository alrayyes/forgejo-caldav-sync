## Why

Coverage was measured locally (`go tool cover`) but never uploaded
anywhere durable or visible — no historical trend, no PR diff
coverage, no badge. Standardizing this across public repos per the
`repo-creation` skill.

## What Changes

- CI uploads `coverage.out` to Codecov on every push/PR.
- A `CODECOV_TOKEN` repo secret is set.
- README gets a Codecov badge next to the existing CI badges.

## Capabilities

### New Capabilities

(none — pipeline/tooling only, no externally observable behavior
change to the service itself)

### Modified Capabilities

(none)

This is pipeline plumbing: `skip_specs: true` is set in
`.openspec.yaml`.

## Impact

`.github/workflows/ci.yml` (coverage upload step), the `CODECOV_TOKEN`
repo secret, `README.md` (badge).
