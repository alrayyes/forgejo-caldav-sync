## Context

See proposal.md - Why. Two systems (Forgejo, a CalDAV server) need to
stay in sync, driven by two different triggers: a real-time webhook
and a periodic pass. Both triggers have to produce identical VTODOs
for the same issue.

## Goals / Non-Goals

**Goals**: keep the mapping/orchestration logic testable without a
real Forgejo or CalDAV server; make the webhook and reconciliation
paths share one code path for "turn this issue into a VTODO",
so they can't drift apart.

**Non-goals**: per-repository calendar splitting (every synced issue
lands in one configured collection); deleting a VTODO on
re-assignment away from a filter (documented limitation, tracked
separately if it becomes worth doing).

## Decisions

- **Hexagonal split**: `internal/sync` is the domain — `Issue`/`Todo`
  types, the mapping between them, assignee filtering, and
  orchestration (`Reconcile`, `HandleIssueEvent`) — against two
  narrow interfaces it owns, `IssueSource` and `CalendarSink`.
  `internal/forgejo` and `internal/caldav` adapt those interfaces to
  the real APIs. `internal/sync` imports neither, which is what makes
  it testable with fakes instead of live servers. Alternative
  considered: a single package doing HTTP-in, mapping, and HTTP-out
  together — rejected because it couples the mapping rules (the part
  that actually needs testing) to two different wire protocols.
- **Reconciliation as the backfill mechanism**: rather than a
  separate one-off import command, the periodic reconciliation pass
  is also what backfills on first run and self-heals a missed
  webhook delivery. One code path instead of three (import, sync,
  self-heal) that could each handle the same issue differently.
- **Deterministic UIDs**: a VTODO's UID is derived from
  `<owner>/<repo>#<number>` rather than a generated identifier, so
  upsert is naturally idempotent — a PUT to the same UID-derived URL
  overwrites in place, with no separate lookup-then-create-or-update
  step.
- **PUT-based CalDAV upsert**: `internal/caldav` doesn't distinguish
  create from update. A PUT to an existing resource overwrites it,
  which makes `Upsert` idempotent for free and avoids a round trip to
  check whether the resource already exists.

## Risks / Trade-offs

- [A missed webhook delivery leaves an issue stale until the next
  reconciliation tick] → Mitigated by `RECONCILE_INTERVAL` defaulting
  to 15 minutes, short enough that staleness is bounded, not by
  guaranteeing webhook delivery.
- [Re-assignment away from an `ASSIGNEE` filter doesn't remove the
  existing VTODO] → Accepted as a known limitation (see proposal);
  deleting on re-assignment is a reasonable next step, not done yet.
- [One CalDAV collection for every synced issue across every
  repository] → Accepted; a per-repo split would need a
  collection-naming scheme this tool doesn't have a use case for yet.

## Migration Plan

N/A — this is the initial implementation, not a migration.
