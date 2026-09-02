## Why

The Release job's `extra_plugins` installed
`@semantic-release/changelog@7.0.0` and `@semantic-release/git@11.0.1`,
but nothing wired them into semantic-release's `plugins` array —
installing a plugin doesn't register it. Confirmed against the v1.0.0
release run: only the default plugin set loaded, no `CHANGELOG.md` was
ever generated, and no release commit was pushed back.

## What Changes

- `.releaserc.json` lists `@semantic-release/changelog` and
  `@semantic-release/git` in `plugins`, after
  `release-notes-generator` and before the `github`/`npm` publish
  steps.
- Every release now generates and commits `CHANGELOG.md`, and pushes a
  `chore(release): <version> [skip ci]` commit.

## Capabilities

### New Capabilities

(none — this fixes the release pipeline's own plugin wiring, it
doesn't change what the service does)

### Modified Capabilities

(none)

This is release-pipeline plumbing: `skip_specs: true` is set in
`.openspec.yaml`.

## Impact

`.releaserc.json` (new), `CHANGELOG.md` (now generated).
