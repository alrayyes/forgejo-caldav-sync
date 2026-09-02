## Why

`rules/go-releases.md` documents release-please (version) + goreleaser
(cross-compiled artefacts) as the pairing for a Go repo on GitHub —
semantic-release is the documented fallback for git.higherlearning.eu
specifically, which has no release-please port. This repo lives on
GitHub and used semantic-release plus a hand-rolled single-arch
`docker build`/`docker push` in `release.yml`, with no goreleaser at
all. Found during a repo-wide audit against
`~/.config/claude/CLAUDE.md` and its `rules/*.md`, filed as a decision
rather than changed blind
([issue #29](https://github.com/alrayyes/forgejo-caldav-sync/issues/29)),
since swapping release tooling touches secrets, the release workflow,
and how the image gets built and tagged.

**Decided: switch to release-please + goreleaser**, matching the
account's other Go repos.

## What Changes

- `release-please-config.json` + `.release-please-manifest.json`
  (starting at `1.0.1`, the version already released) replace
  `.releaserc.json`. release-please opens/updates a release pull
  request instead of committing directly to `main`; merging it — now
  auto-armed the moment it's opened, per `rules/releases.md` — tags
  the release and writes `CHANGELOG.md`.
- `.goreleaser.yml` cross-compiles the binary for linux/darwin ×
  amd64/arm64, and builds + pushes a multi-arch Docker image via
  `dockers_v2` (goreleaser's newer, simpler multi-platform Docker
  config — the older `dockers`/`docker_manifests` pair is deprecated).
- The Dockerfile drops its build stage: goreleaser cross-compiles the
  binary and hands it to `docker build` as part of the build context,
  so the Dockerfile's job is to `COPY $TARGETPLATFORM/forgejo-caldav-sync`
  in, not `go build` it.
- CI's `docker` job and `CONTRIBUTING.md`'s build instructions switch
  from a bare `docker build .` to a goreleaser snapshot build, since a
  bare `docker build .` can no longer prove the image builds on its
  own — there's no binary sitting in the build context without
  goreleaser producing one first.
- Build provenance attestation is preserved (adopted unconditionally
  per `rules/releases.md`) — a small matrix job attests every unique
  image digest goreleaser reports, since `actions/attest-build-provenance`
  takes exactly one digest per invocation.

## Capabilities

### New Capabilities

(none — this changes how the artifact is built and released, not
what the running service does)

### Modified Capabilities

(none)

This is release-pipeline plumbing: `skip_specs: true` is set in
`.openspec.yaml`.

## Impact

`.github/workflows/release.yml` (rewritten), `.github/workflows/ci.yml`
(`docker` job), `Dockerfile`, `CONTRIBUTING.md`, `release-please-config.json`
and `.release-please-manifest.json` (new), `.goreleaser.yml` (new),
`.releaserc.json` (removed). `RELEASE_TOKEN` stays in use — a PR opened
with the default `GITHUB_TOKEN` doesn't trigger `pull_request`-workflows
(including this repo's own required status checks) on itself, so the
release pull request still needs the PAT to get checked and auto-merged
correctly.
