## Why

`rules/go-releases.md` documents release-please (version) + goreleaser
(cross-compiled artefacts) as the pairing for a Go repo on GitHub —
semantic-release is the documented fallback for git.higherlearning.eu
specifically, which has no release-please port. This repo lives on
GitHub and uses semantic-release plus a hand-rolled single-arch
`docker build`/`docker push` in `release.yml`, with no goreleaser at
all. Found during a repo-wide audit against
`~/.config/claude/CLAUDE.md` and its `rules/*.md` — filed as a
decision rather than changed blind, since swapping release tooling
touches secrets, the release workflow, and how the image gets built
and tagged.

**Not yet decided.** See
[issue #29](https://github.com/alrayyes/forgejo-caldav-sync/issues/29)
for the acceptance criteria this proposal has to satisfy either way.

## What Changes

Either:

- Switch to release-please + goreleaser: goreleaser's `dockers:` block
  builds and pushes the image (the Dockerfile only `COPY`s a prebuilt
  binary, it doesn't `go build`), and release-please's manifest starts
  from the current version.

or:

- Keep semantic-release, with a one-line comment in `release.yml`
  recording why (e.g. no multi-arch binaries are shipped today, only
  a single Docker image), so a future audit doesn't re-flag it as
  unexamined drift.

## Capabilities

### New Capabilities

(none — this changes how the artifact is built and released, not
what the running service does)

### Modified Capabilities

(none)

This is release-pipeline plumbing: `skip_specs: true` is set in
`.openspec.yaml`.

## Impact

`release.yml`, possibly a new `.goreleaser.yml` and
`release-please-config.json`/`.release-please-manifest.json`, the
`RELEASE_TOKEN` secret's continued use (or not).
