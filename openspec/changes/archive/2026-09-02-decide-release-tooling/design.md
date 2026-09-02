## Context

See proposal.md - Why. Two separate concerns were previously handled
by one tool (semantic-release) plus a hand-rolled shell script for the
image; the target shape is two tools, each owning one concern, in the
same workflow run.

## Goals / Non-Goals

**Goals**: match `rules/go-releases.md`'s documented pairing; keep the
release pull request itself reviewable and auto-mergeable, not a
direct push; preserve build provenance attestation coverage.

**Non-goals**: SBOM generation (declined per `rules/releases.md` — a
comment in `.goreleaser.yml` notes it's there to flip on), standalone
cosign signing beyond what's already free, a from-source `apt`/AUR/etc
package (this ships a Docker image and Go-installable binaries via
GitHub releases, nothing else).

## Decisions

- **`dockers_v2`, not `dockers`/`docker_manifests`.** `goreleaser
check` flags the older pair as deprecated. `dockers_v2` builds every
  platform for one image in a single buildx invocation instead of one
  `dockers` block per architecture glued back together by a separate
  `docker_manifests` block — confirmed working end-to-end with a real
  `goreleaser release --snapshot --clean --skip=publish` run, not
  just `goreleaser check`.
- **`$TARGETPLATFORM`, not `$TARGETARCH`, in the Dockerfile's `COPY`.**
  goreleaser's own multi-platform build context nests each platform's
  binary under `<os>/<arch>/`, matching Docker buildx's own
  `TARGETPLATFORM` build arg exactly — confirmed live: `$TARGETARCH`
  alone doesn't have the `<os>/` path segment goreleaser's context
  actually uses, and the first attempt at this Dockerfile failed with
  "tried to copy a file that is not available in the build context"
  until corrected.
- **`RELEASE_TOKEN` stays, doesn't move to plain `GITHUB_TOKEN`.**
  Confirmed live (not assumed) that GitHub doesn't run
  `pull_request`-triggered workflows — this repo's own CI included —
  on a pull request opened by the default `GITHUB_TOKEN`. Since this
  repo runs required status checks that gate merging, a release pull
  request opened with the default token would sit with no checks
  ever running on it. A fine-grained PAT belonging the repo owner
  doesn't have that restriction, which is the same reason the
  previous semantic-release setup needed one.
- **Auto-merge armed immediately after release-please opens or
  updates the PR**, in the same job — `rules/releases.md`'s "don't
  wait for a session to notice." `gh pr merge --auto --squash` is
  idempotent to call again on an update, so no state needs tracking
  between runs.
- **Provenance attestation as a matrix job, not a single call.**
  `actions/attest-build-provenance`'s `subject-digest` takes exactly
  one digest; goreleaser's `dist/artifacts.json` reports a "Docker
  Image" entry per tag it pushed (per-architecture tags and the
  merged multi-arch manifest alike), deduplicated by digest and fed
  into a `strategy.matrix` so each real image gets attested exactly
  once regardless of how many tags point at it.
- **CI's `docker` job runs a full goreleaser snapshot, not just a
  Dockerfile syntax check.** A bare `docker build .` can no longer
  prove the image builds — the Dockerfile has nothing to `COPY` without
  goreleaser producing the binary first — so the job now runs
  `goreleaser release --snapshot --clean --skip=publish`, which
  cross-compiles every target and builds (without pushing) both Docker
  platforms, a strictly more thorough check than the one it replaces.

## Risks / Trade-offs

- [`goreleaser release --snapshot` in CI is slower than a bare `docker
build .`] → Accepted; it's still seconds locally (confirmed:
  ~2s for 4 cross-compiled binaries plus 2 Docker images), and it
  catches a cross-compile failure the old check never could.
- [Two release-adjacent GitHub Actions jobs (`release-please`,
  `goreleaser`) instead of one] → Accepted; matches the documented
  split exactly, and `goreleaser` is `if:`-gated on
  `release_created` so it only ever runs on an actual release, not
  every push.

## Migration Plan

1. Land this pull request with everything switched over at once —
   splitting version-decision tooling from artefact-build tooling
   mid-migration would leave a broken intermediate state (release-please
   with no goreleaser to build anything, or vice versa).
2. `.release-please-manifest.json` starts at `1.0.1`, the version
   already tagged, so the next release-please run computes the next
   version from commits after that point rather than restarting
   numbering.
3. No rollback plan beyond reverting the pull request — nothing about
   this migration is destructive to already-published releases or
   images.
