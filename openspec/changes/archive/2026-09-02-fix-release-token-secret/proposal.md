## Why

The `release` job's `actions/checkout` step failed on every run with
`Input required and not supplied: token` — it explicitly passes
`token: ${{ secrets.RELEASE_TOKEN }}`, and that secret didn't exist.
This had been broken since the workflow was first added; merges to
`main` never cut a real release.

## What Changes

- A fine-grained PAT scoped to this repo (Contents: read/write) minted
  and set as the `RELEASE_TOKEN` repo secret.
- The Release workflow's `actions/checkout` step succeeds, and
  `semantic-release` can push a version tag past whatever restricts
  `GITHUB_TOKEN`.

## Capabilities

### New Capabilities

(none — this restores the release pipeline to working order, it
doesn't change what the service itself does)

### Modified Capabilities

(none)

This is pipeline plumbing: `skip_specs: true` is set in
`.openspec.yaml`.

## Impact

The `RELEASE_TOKEN` repo secret, `.github/workflows/release.yml`
(unblocked, no code change needed).

## Note

Minting a PAT is a step only the account owner can do — this change
needed the maintainer's own GitHub account, not the bot's.
