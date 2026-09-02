Retroactive record of already-completed work (issue #1). Every task
below shipped on `main` before this change was documented in OpenSpec.

## 1. Domain

- [x] 1.1 Define `Issue`/`Todo` types and `IssueSource`/`CalendarSink`
      interfaces in `internal/sync` — verified by `go build ./...`
- [x] 1.2 Implement `ToTodo` (issue → VTODO mapping) and `TodoUID`
      (deterministic UID derivation) — verified by
      `internal/sync/sync_test.go`'s `TestToTodo*` and
      `TestToTodoUIDIsStableAcrossRuns`
- [x] 1.3 Implement `FilterByAssignee` (case-insensitive, empty means
      no filter) — verified by
      `TestFilterByAssigneeEmptyKeepsEverything` and
      `TestFilterByAssigneeMatchesCaseInsensitively`
- [x] 1.4 Implement `Reconcile` (list, filter, upsert every match) and
      `HandleIssueEvent` (filter, upsert one) — verified by
      `TestReconcile*` and `TestHandleIssueEvent*`

## 2. Forgejo adapter

- [x] 2.1 Implement paginated issue search against the Forgejo API —
      verified by `internal/forgejo/forgejo_test.go`'s
      `TestListIssuesReturnsMappedIssuesAcrossPages`
- [x] 2.2 Implement webhook payload parsing — verified by
      `TestParseIssueWebhookMapsPayloadToAnIssue` and
      `TestParseIssueWebhookRejectsInvalidJSON`
- [x] 2.3 Implement HMAC-SHA256 signature verification — verified by
      `TestVerifySignatureAcceptsAMatchingHMAC`

## 3. CalDAV adapter

- [x] 3.1 Implement VTODO upsert via PUT, keyed by UID-derived URL —
      verified by `internal/caldav/caldav_test.go`'s
      `TestUpsertPutsAnICSObjectAtTheUIDPath` and
      `TestUpsertMarksDoneIssuesCompleted`
- [x] 3.2 Implement collection creation (MKCALENDAR), idempotent
      against an already-existing collection — verified by
      `TestEnsureCollectionCreatesItWithMkcalendar` and
      `TestEnsureCollectionToleratesAnAlreadyExistingCollection`

## 4. HTTP delivery and composition root

- [x] 4.1 Implement `POST /webhooks/forgejo` (signature check, event
      filter, delegate to `sync.HandleIssueEvent`) and `GET /healthz`
      — verified by `internal/api/api_test.go`
- [x] 4.2 Wire the composition root: config loading, HTTP server,
      reconciliation loop, graceful shutdown on SIGINT/SIGTERM —
      verified by `cmd/forgejo-caldav-sync/main_test.go` and a real
      `docker build .`

## 5. Documentation and delivery

- [x] 5.1 Document every config variable and the install path in
      README.md, with `.env.example` and a Compose file
- [x] 5.2 Green CI (build, test, lint, docker) on the originating pull
      request
