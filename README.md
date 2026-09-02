# forgejo-caldav-sync

[![CI](https://github.com/alrayyes/forgejo-caldav-sync/actions/workflows/ci.yml/badge.svg)](https://github.com/alrayyes/forgejo-caldav-sync/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/alrayyes/forgejo-caldav-sync/graph/badge.svg)](https://codecov.io/gh/alrayyes/forgejo-caldav-sync)
[![License: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/alrayyes/forgejo-caldav-sync.svg)](https://pkg.go.dev/github.com/alrayyes/forgejo-caldav-sync)

Syncs Forgejo issues into a CalDAV calendar as VTODOs — open issues appear
as tasks, closed issues get marked done, and there's nothing to re-invoke:
start it once and leave it running.

Forgejo has no built-in CalDAV or ICS export
([forgejo/forgejo#4887](https://codeberg.org/forgejo/forgejo/issues/4887) is
still an open, unimplemented feature request), so this bridges the gap
instead of waiting for one.

## How it works

Two things run at once:

- **A webhook receiver** at `/webhooks/forgejo`. Point a Forgejo webhook
  (repository-level, or instance-wide via Admin → System Webhooks, to cover
  every repo without configuring each one) at it for the "Issues" event,
  and an issue being opened, edited, closed or reopened updates its VTODO
  within seconds of the event arriving.
- **A periodic reconciliation pass**, on `RECONCILE_INTERVAL`. It lists
  every issue visible to the configured token and upserts each one, the
  same way the webhook path does for a single issue. This is what backfills
  older issues on first run — there's no separate import command, because
  reconciling and backfilling are the same operation — and it self-heals
  anything a missed webhook delivery left out of sync.

Every VTODO's `UID` is derived deterministically from the issue it
represents (`<owner>/<repo>#<number>`), so re-syncing the same issue always
updates the same CalDAV resource instead of creating a duplicate. A closed
issue's VTODO gets `STATUS:COMPLETED` and a `COMPLETED` timestamp; an open
one gets `STATUS:NEEDS-ACTION`. If the issue has a due date, it becomes the
VTODO's `DUE`.

By default every issue on the instance syncs, regardless of assignee — set
`ASSIGNEE` to narrow that down to one person's tickets, which is what makes
this usable by anyone running a Forgejo instance, not only for a single
account's full ticket list.

## Requirements

- A Forgejo instance with an API token that can read the issues you want
  synced, and permission to add a webhook (repository or instance admin,
  depending on scope).
- A CalDAV server reachable from wherever this runs, and a collection (task
  list) URL on it — this tool creates the collection on first run if it
  doesn't already exist.
- Docker (or Go 1.27+, if building from source).

## Installation

Pull the published image — a merge to `main` carrying a `feat`/`fix` commit
releases a new version to `ghcr.io`, tagged with both the version and
`latest`:

```sh
mkdir forgejo-caldav-sync && cd forgejo-caldav-sync
curl -O https://raw.githubusercontent.com/alrayyes/forgejo-caldav-sync/main/compose.yaml
curl -O https://raw.githubusercontent.com/alrayyes/forgejo-caldav-sync/main/.env.example
cp .env.example .env # fill in your Forgejo and CalDAV details
docker run -d --restart unless-stopped \
  --env-file .env -p 8080:8080 \
  ghcr.io/alrayyes/forgejo-caldav-sync:latest
```

Or build from source:

```sh
git clone https://github.com/alrayyes/forgejo-caldav-sync.git
cd forgejo-caldav-sync
cp .env.example .env # fill in your Forgejo and CalDAV details
docker compose up -d
```

Then, on your Forgejo instance, add a webhook pointed at
`http://<this-service>:8080/webhooks/forgejo` for the "Issues" event, with
the same secret as `FORGEJO_WEBHOOK_SECRET`.

## Configuration

Every setting can come from a flag, an environment variable or a config
file — in that order of precedence, so a flag always wins and a config
file is the fallback underneath everything else.

| Setting                 | Flag                       | Environment variable     | Required | Default                  |
| ----------------------- | -------------------------- | ------------------------ | -------- | ------------------------ |
| Forgejo base URL        | `--forgejo-base-url`       | `FORGEJO_BASE_URL`       | yes      | —                        |
| Forgejo API token       | `--forgejo-token`          | `FORGEJO_TOKEN`          | yes      | —                        |
| Forgejo webhook secret  | `--forgejo-webhook-secret` | `FORGEJO_WEBHOOK_SECRET` | yes      | —                        |
| CalDAV collection URL   | `--caldav-url`             | `CALDAV_URL`             | yes      | —                        |
| CalDAV username         | `--caldav-username`        | `CALDAV_USERNAME`        | yes      | —                        |
| CalDAV password         | `--caldav-password`        | `CALDAV_PASSWORD`        | yes      | —                        |
| Assignee filter         | `--assignee`               | `ASSIGNEE`               | no       | unset — sync every issue |
| HTTP listen address     | `--addr`                   | `ADDR`                   | no       | `:8080`                  |
| Reconciliation interval | `--reconcile-interval`     | `RECONCILE_INTERVAL`     | no       | `15m`                    |

`CALDAV_URL`/`--caldav-url` is the collection (task list) URL, not the
server root — for a Baïkal instance, that's typically
`https://your-dav-server/dav.php/calendars/<user>/<collection>/`; check
whatever your CalDAV server's own URL layout is.

Running the Docker image (`.env` + `docker-compose`/`docker run
--env-file`) is still the documented path — environment variables keep
working exactly as before. `forgejo-caldav-sync init` writes a starter
config file (at `$XDG_CONFIG_HOME/forgejo-caldav-sync/config.yaml`, or
pass `--config <path>` to use a different one) for anyone running the
binary directly instead. A run with no config file and no relevant
environment variable set, at an interactive terminal, offers to write one
for you.

## Usage

Once running, it logs structured JSON:

```json
{ "level": "INFO", "msg": "reconciled", "synced": 3 }
```

It shuts down cleanly on `SIGTERM`/`SIGINT` — `docker compose down` or
`docker stop` are safe. `GET /healthz` and the image's own `HEALTHCHECK`
(which calls it internally — the distroless base has no shell or curl for a
CMD-SHELL check to run) report the process is up, not that Forgejo or the
CalDAV server are reachable: a real outage on either side should keep
retrying on the next reconciliation pass, not get "fixed" by a supervisor
restarting a container that was never actually stuck.

## Known limitations

- **Re-assignment doesn't remove a VTODO.** If `ASSIGNEE` is set and an
  issue is re-assigned away from that user, the webhook path leaves its
  existing VTODO alone rather than deleting it — the next reconciliation
  pass doesn't touch it either, since it only ever upserts what currently
  matches. Deleting a VTODO a person no longer owns is a reasonable next
  step; it just isn't done yet.
- **One CalDAV collection.** Every synced issue, across every repository
  that matches, lands in the single collection `CALDAV_URL` points at —
  there's no per-repo calendar split.

## Testing

```sh
go test ./...
```

Every adapter is tested against an `httptest` server standing in for the
real Forgejo API or CalDAV server — there's no live-service test layer.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[GPL-3.0](LICENSE).
