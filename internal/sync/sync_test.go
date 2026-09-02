package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
	"github.com/stretchr/testify/require"
)

var (
	errBoom              = errors.New("boom")
	errCalDAVUnreachable = errors.New("caldav unreachable")
)

func dueAt(t *testing.T, s string) *time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)

	return &tm
}

func TestToTodoOpenIssue(t *testing.T) {
	t.Parallel()

	issue := sync.Issue{
		RepoFullName: "alice/widgets",
		Number:       42,
		Title:        "Widget falls over",
		Body:         "Steps to reproduce...",
		URL:          "https://forge.example.com/alice/widgets/issues/42",
		State:        "open",
		Assignees:    []string{"bob"},
		DueDate:      dueAt(t, "2026-09-01T00:00:00Z"),
	}

	todo := sync.ToTodo(issue)

	require.Equal(t, "forgejo-caldav-sync-alice-widgets-42", todo.UID)
	require.Equal(t, "alice/widgets#42: Widget falls over", todo.Summary)
	require.Contains(t, todo.Description, "Steps to reproduce...")
	require.Contains(t, todo.Description, issue.URL)
	require.Equal(t, issue.URL, todo.URL)
	require.False(t, todo.Done)
	require.Equal(t, issue.DueDate, todo.Due)
	require.Nil(t, todo.Completed)
}

func TestToTodoClosedIssue(t *testing.T) {
	t.Parallel()

	closedAt := dueAt(t, "2026-08-20T12:00:00Z")
	issue := sync.Issue{
		RepoFullName: "alice/widgets",
		Number:       7,
		Title:        "Old bug",
		State:        "closed",
		ClosedAt:     closedAt,
	}

	todo := sync.ToTodo(issue)

	require.True(t, todo.Done)
	require.Equal(t, closedAt, todo.Completed)
}

func TestToTodoUIDIsStableAcrossRuns(t *testing.T) {
	t.Parallel()

	issue := sync.Issue{RepoFullName: "alice/widgets", Number: 42, Title: "x"}

	first := sync.ToTodo(issue)
	second := sync.ToTodo(issue)

	require.Equal(t, first.UID, second.UID)
}

func TestFilterByAssigneeEmptyKeepsEverything(t *testing.T) {
	t.Parallel()

	issues := []sync.Issue{
		{Number: 1, Assignees: []string{"alice"}},
		{Number: 2, Assignees: nil},
	}

	require.Equal(t, issues, sync.FilterByAssignee(issues, ""))
}

func TestFilterByAssigneeMatchesCaseInsensitively(t *testing.T) {
	t.Parallel()

	issues := []sync.Issue{
		{Number: 1, Assignees: []string{"Alice"}},
		{Number: 2, Assignees: []string{"bob"}},
		{Number: 3, Assignees: []string{"alice", "carol"}},
	}

	filtered := sync.FilterByAssignee(issues, "alice")

	require.Len(t, filtered, 2)
	require.Equal(t, int64(1), filtered[0].Number)
	require.Equal(t, int64(3), filtered[1].Number)
}

type fakeSource struct {
	issues []sync.Issue
	err    error
}

func (f fakeSource) ListIssues(_ context.Context) ([]sync.Issue, error) {
	return f.issues, f.err
}

type fakeSink struct {
	upserted []sync.Todo
	err      error
}

func (f *fakeSink) Upsert(_ context.Context, todo sync.Todo) error {
	if f.err != nil {
		return f.err
	}
	f.upserted = append(f.upserted, todo)

	return nil
}

func TestReconcileUpsertsEveryMatchingIssue(t *testing.T) {
	t.Parallel()

	src := fakeSource{issues: []sync.Issue{
		{RepoFullName: "alice/widgets", Number: 1, Assignees: []string{"alice"}},
		{RepoFullName: "alice/widgets", Number: 2, Assignees: []string{"bob"}},
	}}
	sink := &fakeSink{}

	n, err := sync.Reconcile(context.Background(), src, sink, "alice")

	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Len(t, sink.upserted, 1)
	require.Equal(t, "forgejo-caldav-sync-alice-widgets-1", sink.upserted[0].UID)
}

func TestReconcileNoFilterSyncsEverything(t *testing.T) {
	t.Parallel()

	src := fakeSource{issues: []sync.Issue{
		{RepoFullName: "alice/widgets", Number: 1},
		{RepoFullName: "alice/widgets", Number: 2},
	}}
	sink := &fakeSink{}

	n, err := sync.Reconcile(context.Background(), src, sink, "")

	require.NoError(t, err)
	require.Equal(t, 2, n)
}

func TestReconcilePropagatesSourceError(t *testing.T) {
	t.Parallel()

	src := fakeSource{err: errBoom}
	sink := &fakeSink{}

	_, err := sync.Reconcile(context.Background(), src, sink, "")

	require.Error(t, err)
}

func TestReconcileStopsOnFirstSinkError(t *testing.T) {
	t.Parallel()

	src := fakeSource{issues: []sync.Issue{
		{RepoFullName: "alice/widgets", Number: 1},
		{RepoFullName: "alice/widgets", Number: 2},
	}}
	sink := &fakeSink{err: errCalDAVUnreachable}

	_, err := sync.Reconcile(context.Background(), src, sink, "")

	require.Error(t, err)
}

func TestHandleIssueEventUpsertsWhenNoFilter(t *testing.T) {
	t.Parallel()

	sink := &fakeSink{}
	issue := sync.Issue{RepoFullName: "alice/widgets", Number: 5, Assignees: []string{"bob"}}

	err := sync.HandleIssueEvent(context.Background(), sink, "", issue)

	require.NoError(t, err)
	require.Len(t, sink.upserted, 1)
}

func TestHandleIssueEventSkipsWhenFilteredOut(t *testing.T) {
	t.Parallel()

	sink := &fakeSink{}
	issue := sync.Issue{RepoFullName: "alice/widgets", Number: 5, Assignees: []string{"bob"}}

	err := sync.HandleIssueEvent(context.Background(), sink, "alice", issue)

	require.NoError(t, err)
	require.Empty(t, sink.upserted)
}

func TestHandleIssueEventUpsertsWhenAssigneeMatches(t *testing.T) {
	t.Parallel()

	sink := &fakeSink{}
	issue := sync.Issue{RepoFullName: "alice/widgets", Number: 5, Assignees: []string{"alice"}}

	err := sync.HandleIssueEvent(context.Background(), sink, "alice", issue)

	require.NoError(t, err)
	require.Len(t, sink.upserted, 1)
}
