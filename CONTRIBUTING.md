# Contributing

## Toolchain

- Go 1.27+
- Docker, for building the image and running the Dockerfile lint
- [Bun](https://bun.sh) 1.3.14, for the documentation tooling (Prettier,
  markdownlint, commitlint) — nothing here is a JavaScript project

```sh
go mod download
bun install
bun run prepare # installs the git hooks (lefthook)
```

## Building

```sh
go build ./cmd/forgejo-caldav-sync
docker build -t forgejo-caldav-sync .
```

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
docker run --rm -i hadolint/hadolint:v2.15.1@sha256:32dac94127fd60b7b7e3fbfc65e1383b9b5e25c9bfd7b8536de7a539fe68a12d < Dockerfile
bun run format:check      # prettier, markdown and yaml
bun run lint:md           # markdownlint
bun run lint:api          # redocly, against api/openapi.yaml
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
`internal/sync`, same as `cmd/forgejo-caldav-sync/main.go`'s reconciliation
loop does on the outbound side. Nothing in `internal/sync` imports any of
the other three; that's what makes it testable without a Forgejo or CalDAV
instance up.

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
branch. [semantic-release](https://semantic-release.gitbook.io/) reads the
Conventional Commits on `main` and, on the next push, tags the release,
writes `CHANGELOG.md`, creates the GitHub release, and builds and pushes the
image to `ghcr.io/alrayyes/forgejo-caldav-sync`, tagged with both the
version and `latest`. Nobody picks a version by hand.
