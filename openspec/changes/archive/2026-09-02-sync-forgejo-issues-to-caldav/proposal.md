## Why

Forgejo has no built-in CalDAV or ICS export
([forgejo/forgejo#4887](https://codeberg.org/forgejo/forgejo/issues/4887)
is still an open, unimplemented feature request), and nothing
third-party bridges Forgejo issues into CalDAV. Someone running a
self-hosted Forgejo instance has no way to see or manage their tickets
from a CalDAV-aware calendar or task app without checking the forge
directly.

## What Changes

- A webhook receiver at `POST /webhooks/forgejo` that upserts a
  matching CalDAV VTODO within seconds of a Forgejo "issues" event
  (opened, edited, closed, reopened).
- A periodic reconciliation pass that lists every issue visible to the
  configured token and upserts each one — this is also what backfills
  older issues on first run, since it's the same operation.
- Deterministic VTODO UIDs derived from `<owner>/<repo>#<number>`, so
  re-syncing an issue always updates the same CalDAV resource instead
  of creating a duplicate.
- Optional `ASSIGNEE` filtering, so the tool works for a single
  person's ticket list or for an entire instance depending on
  configuration.
- HMAC-SHA256 webhook signature verification, rejecting anything with
  a missing or incorrect signature before it reaches the calendar.

## Capabilities

### New Capabilities

- `issue-sync`: mapping a Forgejo issue to a CalDAV VTODO, filtering
  by assignee, and the two ways sync happens (a real-time webhook
  event, and a periodic reconciliation pass that also backfills and
  self-heals).

### Modified Capabilities

(none — this is the initial capability)

## Impact

Introduces `internal/sync` (the domain: mapping and orchestration
against two narrow interfaces), `internal/forgejo` (Forgejo API
adapter: issue search, webhook parsing, signature verification),
`internal/caldav` (CalDAV adapter: VTODO upsert via PUT, collection
creation), `internal/api` (inbound HTTP: the webhook handler and
`/healthz`), and `cmd/forgejo-caldav-sync` (the composition root: wires
the two adapters together, runs the HTTP server and the reconciliation
loop).
