## 1. internal/sync

- [x] 1.1 Split `TestToTodo`'s open/closed-issue mapping into
      `TestToTodoOpenIssue` and `TestToTodoClosedIssue`, one `t.Run`
      subtest per field — verified by `go test -race ./internal/sync/...`
- [x] 1.2 Split `TestReconcile` into `TestReconcileAssigneeFilter` and
      `TestReconcileErrors` — verified the same way
- [x] 1.3 Collapse `TestFilterByAssigneeMatchesCaseInsensitively`'s
      length + per-index checks into one assertion on the matched issue
      numbers — verified the same way

## 2. internal/api

- [x] 2.1 Extract a `postWebhook` helper and split every
      status-code+upsert-state test into two subtests — verified by
      `go test -race ./internal/api/...`

## 3. internal/caldav

- [x] 3.1 Extract a `recordUpsert` helper and split the PUT-request and
      VTODO-body assertions into one subtest per field — verified by
      `go test -race ./internal/caldav/...`
- [x] 3.2 Convert the already-exists status-code loop into named
      `t.Run` subtests — verified the same way

## 4. internal/forgejo

- [x] 4.1 Split `TestListIssuesReturnsMappedIssuesAcrossPages` into
      `TestListIssuesPagination` and `TestListIssuesMapping` — verified
      by `go test -race ./internal/forgejo/...`
- [x] 4.2 Split `TestVerifySignatureAcceptsAMatchingHMAC`'s three
      scenarios (valid, incorrect, wrong secret) into three tests —
      verified the same way
- [x] 4.3 Split `TestParseIssueWebhookMapsPayloadToAnIssue`'s six
      checks into subtests — verified the same way

## 5. internal/config

- [x] 5.1 Split `TestLoadWithOnlyRequiredVariables` and
      `TestLoadHonorsOptionalOverrides` into per-field subtests —
      verified by `go test -race ./internal/config/...`

## 6. Verification

- [x] 6.1 `go test -race ./...` green, `golangci-lint run ./...` clean
      (funlen and testifylint findings from the first pass fixed by
      splitting oversized parent tests and comparing two distinct
      variables instead of two identical call expressions)
- [x] 6.2 Coverage unchanged: 73.6% total, same per-package numbers as
      before the restructuring
