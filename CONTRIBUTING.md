# Contributing

## Toolchain

- Go 1.27+
- Docker, for building the image and running the Dockerfile lint
- [goreleaser](https://goreleaser.com/) 2.x, for building the binary and
  the Docker image the same way the release pipeline does
- [Bun](https://bun.sh) 1.3.14, for the documentation tooling (Prettier,
  markdownlint, commitlint) — nothing here is a JavaScript project

```sh
go mod download
bun install
bun run prepare # installs the git hooks (lefthook)
```

## Building

```sh
go build ./cmd/forgejo-caldav-sync   # for local testing — not what ships
goreleaser build --single-target --snapshot --clean
```

The Docker image isn't built with a plain `docker build .` anymore —
goreleaser cross-compiles the binary first and the Dockerfile only
`COPY`s it in, so building the image the same way the release does needs
goreleaser too:

```sh
goreleaser release --snapshot --clean --skip=publish
```

That builds every release target (Linux and macOS, amd64 and arm64) and
both Docker image platforms, without pushing anything.

## Testing

```sh
go test ./...                                            # no external services needed
go test -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

Every adapter (`internal/forgejo`, `internal/caldav`) is tested against an
`httptest` server standing in for the real thing, not a live Forgejo or
Baïkal instance — the mapping/orchestration logic in `internal/sync` is
tested without HTTP at all, against fakes implementing its two interfaces.

## Linting

```sh
golangci-lint run ./...
golangci-lint fmt ./...   # gofumpt + goimports
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
go mod tidy -diff
docker run --rm -i hadolint/hadolint:v2.15.1@sha256:32dac94127fd60b7b7e3fbfc65e1383b9b5e25c9bfd7b8536de7a539fe68a12d < Dockerfile
goreleaser check
bun run format:check      # prettier, markdown and yaml
bun run lint:md           # markdownlint
bun run lint:api          # redocly, against api/openapi.yaml
bun run audit             # bun audit, against the pinned dev-tooling deps
```

CI runs exactly these commands — see `.github/workflows/ci.yml`. The git
hooks in `lefthook.yml` run the fast subset on commit and the rest on push,
so a red pipeline should never be a surprise.

## How it fits together

`internal/sync` is the domain: `Issue`/`Todo` types, the mapping between
them, assignee filtering, and orchestration (`Reconcile`, `HandleIssueEvent`)
against two narrow interfaces — `IssueSource` and `CalendarSink`. It knows
nothing about HTTP or wire formats, and its tests use fakes, not real
servers.

`internal/forgejo` and `internal/caldav` adapt those interfaces to the real
Forgejo API and a real CalDAV server, respectively. `internal/api` is the
inbound HTTP adapter — it decodes a webhook delivery and calls into
`internal/sync`, same as `cmd/forgejo-caldav-sync/serve.go`'s reconciliation
loop does on the outbound side. Nothing in `internal/sync` imports any of
the other three; that's what makes it testable without a Forgejo or CalDAV
instance up.

`internal/config` is just the `Config` struct and its `Validate()` — no
cobra or viper import, so it stays framework-free and trivially testable.
Resolving that struct from flags, environment and a config file is
`cmd/forgejo-caldav-sync`'s job, since that layering is inherently tied to
the command-line surface: `root.go` builds the flag set and viper wiring
(`resolveConfig`), `serve.go` is the composition root for the default
command, `init.go` writes a starter config file, and `healthcheck.go` is
the hidden subcommand the Dockerfile's `HEALTHCHECK` calls.

## The contract

`api/openapi.yaml` describes the two HTTP endpoints and is handwritten, not
generated from the handlers. `redocly lint` checks the document is valid
OpenAPI; nothing yet checks the handler still matches it.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/):
`type(scope): description`, types `feat`/`fix`/`docs`/`style`/`refactor`/
`perf`/`test`/`build`/`ci`/`chore`/`revert`. Subject under 50 characters,
lowercase, no trailing full stop. commitlint enforces the shape at
commit-msg and again in CI; the length and case rules are tighter than what
it checks, so hold to them anyway.

## Branching, review, and release

Every change goes through a pull request — nothing is pushed straight to
`main`.

The pull request **title** has to be a valid Conventional Commit too —
commitlint checks it against the base commit range in CI. A squash merge
defaults its commit message to the pull request title, so this is the only
check standing between a badly titled pull request and a bad message on
`main`.

Once a pull request's checks are green, squash-merge it and delete the
branch. [release-please](https://github.com/googleapis/release-please)
reads the Conventional Commits on `main` and keeps a release pull request
open with the next version and `CHANGELOG.md` entry; merging that pull
request (auto-merged the moment its own checks pass) tags the release and
creates the GitHub release. That tag then drives
[goreleaser](https://goreleaser.com/), which cross-compiles the binary for
Linux and macOS (amd64 and arm64), and builds and pushes a multi-arch
Docker image to `ghcr.io/alrayyes/forgejo-caldav-sync`, tagged with both
the version and `latest`. Nobody picks a version by hand.
