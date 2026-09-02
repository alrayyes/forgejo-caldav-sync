# issue-sync Specification

## Purpose

Keeps a CalDAV task list in sync with the state of every issue on a
Forgejo instance (or a filtered subset), so tickets are visible and
manageable from any CalDAV-aware calendar or task app.

## Requirements

### Requirement: Map an issue to a VTODO

The system SHALL derive a VTODO deterministically from a Forgejo
issue, so re-syncing the same issue always resolves to the same
CalDAV resource instead of creating a duplicate.

#### Scenario: Deterministic UID

- **WHEN** the same issue is mapped to a VTODO more than once
- **THEN** the resulting UID is identical every time, derived from
  `<owner>/<repo>#<number>`

#### Scenario: Open issue maps to an actionable VTODO

- **WHEN** an open issue is mapped
- **THEN** the VTODO has `STATUS:NEEDS-ACTION`, the issue title as
  `SUMMARY`, the issue body and URL in `DESCRIPTION`, and `DUE` set
  from the issue's due date if it has one

#### Scenario: Closed issue maps to a completed VTODO

- **WHEN** a closed issue is mapped
- **THEN** the VTODO has `STATUS:COMPLETED` and `COMPLETED` set from
  the issue's closed timestamp

### Requirement: Real-time sync via webhook

The system SHALL upsert a matching VTODO within seconds of a Forgejo
"issues" webhook event (opened, edited, closed, reopened), using the
same mapping as reconciliation.

#### Scenario: Issue event upserts its VTODO

- **WHEN** a Forgejo "issues" webhook event arrives for an issue
  matching the configured assignee filter (or no filter is set)
- **THEN** that issue's VTODO is created or overwritten in the
  configured CalDAV collection

#### Scenario: Non-issue events are accepted and ignored

- **WHEN** a webhook delivery arrives for an event other than
  "issues"
- **THEN** the system responds successfully without writing to the
  calendar, since Forgejo has no per-event webhook URL and the
  endpoint must tolerate anything configured to be sent to it

#### Scenario: Re-assignment away from the filter leaves the VTODO alone

- **WHEN** an issue is re-assigned away from a configured `ASSIGNEE`
  filter
- **THEN** the webhook path leaves that issue's existing VTODO
  untouched rather than deleting it — deletion is a known limitation,
  not yet implemented

### Requirement: Periodic reconciliation, backfill and self-heal

The system SHALL periodically list every issue visible to the
configured token and upsert each one matching the assignee filter,
using the same mapping as the webhook path.

#### Scenario: First run backfills existing issues

- **WHEN** the service starts against an instance with existing
  issues and no prior state
- **THEN** the first reconciliation pass upserts every matching
  issue as a VTODO — there is no separate one-off import command,
  since reconciling and backfilling are the same operation

#### Scenario: A missed webhook delivery self-heals

- **WHEN** a webhook delivery is missed or fails to arrive
- **THEN** the next reconciliation pass still upserts that issue
  correctly, since reconciliation re-derives every VTODO from
  current issue state rather than tracking deltas

### Requirement: Assignee filtering

The system SHALL restrict sync to issues assigned to a configured
username when one is set, and sync every issue when it is not.

#### Scenario: No filter configured

- **WHEN** `ASSIGNEE` is left unset
- **THEN** every issue on the instance is synced, regardless of
  assignee

#### Scenario: Filter configured

- **WHEN** `ASSIGNEE` is set to a username
- **THEN** only issues with that user among the assignees are
  synced, matched case-insensitively since Forgejo usernames aren't
  case sensitive

### Requirement: Webhook signature verification

The system SHALL verify every inbound webhook delivery's HMAC-SHA256
signature before acting on it.

#### Scenario: Missing or incorrect signature is rejected

- **WHEN** a webhook request arrives with a missing or incorrect
  signature
- **THEN** the request is rejected and nothing is written to the
  calendar
