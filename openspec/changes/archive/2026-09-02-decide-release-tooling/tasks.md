## 1. release-please

- [x] 1.1 Add `release-please-config.json` (`release-type: go`) and
      `.release-please-manifest.json` starting at `1.0.1` — verified
      by `googleapis/release-please-action`'s config-file/manifest-file
      inputs pointing at them
- [x] 1.2 Remove `.releaserc.json` (semantic-release config) —
      verified by `grep -r semantic-release` finding only historical
      OpenSpec archive entries and the tool being removed
- [x] 1.3 Wire the release pull request to auto-merge itself the
      moment it's opened or updated — verified by the workflow step
      gated on `prs_created == 'true'`, reading the PR number from
      `fromJSON(steps.release.outputs.pr).number`

## 2. goreleaser

- [x] 2.1 Add `.goreleaser.yml`: cross-compile linux/darwin ×
      amd64/arm64, archives, checksums, changelog disabled (release-please
      already wrote it), `release.mode: keep-existing` — verified by
      `goreleaser check`
- [x] 2.2 Configure `dockers_v2` for a multi-platform image build in
      one entry — verified by a real
      `goreleaser release --snapshot --clean --skip=publish` run
      producing both `linux/amd64` and `linux/arm64` images and
      `goreleaser check` reporting no deprecation warnings
- [x] 2.3 Inject the version via `-X main.version={{.Version}}` —
      verified by running the snapshot-built image with `--version`
      and confirming it printed the snapshot version, not `dev`

## 3. Dockerfile

- [x] 3.1 Drop the build stage; `COPY $TARGETPLATFORM/forgejo-caldav-sync`
      instead of `go build` — verified by the snapshot build
      succeeding for both platforms and
      `docker run --rm <image> --version` printing the expected version
- [x] 3.2 Confirm hadolint still passes — verified by running it
      against the rewritten Dockerfile directly

## 4. CI

- [x] 4.1 Replace the `docker` job's bare `docker build .` with a
      goreleaser snapshot build (`--snapshot --clean --skip=publish`),
      since the Dockerfile has nothing to `COPY` without one —
      verified by running the same command locally
- [x] 4.2 Add `goreleaser check` to `CONTRIBUTING.md`'s linting
      command list

## 5. Release workflow

- [x] 5.1 Rewrite `release.yml`: a `release-please` job producing
      `release_created`/`tag_name` outputs, a `goreleaser` job gated
      on `release_created == 'true'` (same workflow run — a tag pushed
      with the default `GITHUB_TOKEN` starts no further workflow), and
      an `attest` matrix job attesting every unique image digest —
      verified by `actionlint`
- [x] 5.2 Keep `RELEASE_TOKEN` for the `release-please` step —
      confirmed live that a PR opened with the default `GITHUB_TOKEN`
      doesn't trigger this repo's own required-status-check workflows
      on itself

## 6. Documentation

- [x] 6.1 `CONTRIBUTING.md`'s Building and Branching/release sections
      describe the new build and release flow
- [x] 6.2 `openspec/config.yaml`'s project context updated (no more
      "semantic-release", now "release-please decides the version...
      goreleaser cross-compiles...")

## 7. Verification

- [x] 7.1 `go build ./...`, `go test -race ./...`,
      `golangci-lint run ./...` all still green (this change touches
      no Go source)
- [x] 7.2 `goreleaser check` clean, no deprecation warnings
- [x] 7.3 `actionlint` clean on `release.yml` and `ci.yml`
- [x] 7.4 A real `goreleaser release --snapshot --clean --skip=publish`
      run succeeds end-to-end: 4 cross-compiled binaries, 4 archives,
      checksums, 2 Docker images (amd64 + arm64), each running
      correctly with the right embedded version
