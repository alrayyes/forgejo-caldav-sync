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

func openIssueFixture(t *testing.T) sync.Issue {
	t.Helper()

	return sync.Issue{
		RepoFullName: "alice/widgets",
		Number:       42,
		Title:        "Widget falls over",
		Body:         "Steps to reproduce...",
		URL:          "https://forge.example.com/alice/widgets/issues/42",
		State:        "open",
		Assignees:    []string{"bob"},
		DueDate:      dueAt(t, "2026-09-01T00:00:00Z"),
	}
}

func TestToTodoOpenIssue(t *testing.T) {
	t.Parallel()

	issue := openIssueFixture(t)
	todo := sync.ToTodo(issue)

	t.Run("derives UID from repo and issue number", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "forgejo-caldav-sync-alice-widgets-42", todo.UID)
	})

	t.Run("summary combines repo, number and title", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "alice/widgets#42: Widget falls over", todo.Summary)
	})

	t.Run("description contains the issue body", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, todo.Description, issue.Body)
	})

	t.Run("description contains the issue URL", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, todo.Description, issue.URL)
	})

	t.Run("URL matches the issue URL", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, issue.URL, todo.URL)
	})

	t.Run("is not done", func(t *testing.T) {
		t.Parallel()
		require.False(t, todo.Done)
	})

	t.Run("due matches the issue due date", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, issue.DueDate, todo.Due)
	})

	t.Run("has no completed time", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, todo.Completed)
	})
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

	t.Run("is done", func(t *testing.T) {
		t.Parallel()
		require.True(t, todo.Done)
	})

	t.Run("completed matches the issue's closed time", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, closedAt, todo.Completed)
	})
}

func TestToTodoUIDIsStableAcrossRuns(t *testing.T) {
	t.Parallel()

	issue := openIssueFixture(t)
	first := sync.ToTodo(issue)
	second := sync.ToTodo(issue)

	require.Equal(t, first.UID, second.UID)
}

func TestFilterByAssignee(t *testing.T) {
	t.Parallel()

	t.Run("empty username keeps every issue", func(t *testing.T) {
		t.Parallel()
		issues := []sync.Issue{
			{Number: 1, Assignees: []string{"alice"}},
			{Number: 2, Assignees: nil},
		}

		require.Equal(t, issues, sync.FilterByAssignee(issues, ""))
	})

	t.Run("matches a username case-insensitively", func(t *testing.T) {
		t.Parallel()
		issues := []sync.Issue{
			{Number: 1, Assignees: []string{"Alice"}},
			{Number: 2, Assignees: []string{"bob"}},
			{Number: 3, Assignees: []string{"alice", "carol"}},
		}

		filtered := sync.FilterByAssignee(issues, "alice")

		numbers := make([]int64, len(filtered))
		for i, issue := range filtered {
			numbers[i] = issue.Number
		}
		require.Equal(t, []int64{1, 3}, numbers)
	})
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

func TestReconcileAssigneeFilter(t *testing.T) {
	t.Parallel()

	t.Run("upserts only issues matching the assignee filter", func(t *testing.T) {
		t.Parallel()
		src := fakeSource{issues: []sync.Issue{
			{RepoFullName: "alice/widgets", Number: 1, Assignees: []string{"alice"}},
			{RepoFullName: "alice/widgets", Number: 2, Assignees: []string{"bob"}},
		}}
		sink := &fakeSink{}

		n, err := sync.Reconcile(context.Background(), src, sink, "alice")

		require.NoError(t, err)
		require.Equal(t, 1, n)
	})

	t.Run("upserts the matching issue's mapped UID", func(t *testing.T) {
		t.Parallel()
		src := fakeSource{issues: []sync.Issue{
			{RepoFullName: "alice/widgets", Number: 1, Assignees: []string{"alice"}},
			{RepoFullName: "alice/widgets", Number: 2, Assignees: []string{"bob"}},
		}}
		sink := &fakeSink{}

		_, err := sync.Reconcile(context.Background(), src, sink, "alice")

		require.NoError(t, err)
		require.Equal(t, []string{"forgejo-caldav-sync-alice-widgets-1"}, upsertedUIDs(sink))
	})

	t.Run("no filter syncs every issue", func(t *testing.T) {
		t.Parallel()
		src := fakeSource{issues: []sync.Issue{
			{RepoFullName: "alice/widgets", Number: 1},
			{RepoFullName: "alice/widgets", Number: 2},
		}}
		sink := &fakeSink{}

		n, err := sync.Reconcile(context.Background(), src, sink, "")

		require.NoError(t, err)
		require.Equal(t, 2, n)
	})
}

func TestReconcileErrors(t *testing.T) {
	t.Parallel()

	t.Run("propagates a source error", func(t *testing.T) {
		t.Parallel()
		src := fakeSource{err: errBoom}
		sink := &fakeSink{}

		_, err := sync.Reconcile(context.Background(), src, sink, "")

		require.Error(t, err)
	})

	t.Run("stops on the first sink error", func(t *testing.T) {
		t.Parallel()
		src := fakeSource{issues: []sync.Issue{
			{RepoFullName: "alice/widgets", Number: 1},
			{RepoFullName: "alice/widgets", Number: 2},
		}}
		sink := &fakeSink{err: errCalDAVUnreachable}

		_, err := sync.Reconcile(context.Background(), src, sink, "")

		require.Error(t, err)
	})
}

func upsertedUIDs(sink *fakeSink) []string {
	uids := make([]string, len(sink.upserted))
	for i, todo := range sink.upserted {
		uids[i] = todo.UID
	}

	return uids
}

func TestHandleIssueEvent(t *testing.T) {
	t.Parallel()

	t.Run("upserts when no filter is set", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		issue := sync.Issue{RepoFullName: "alice/widgets", Number: 5, Assignees: []string{"bob"}}

		err := sync.HandleIssueEvent(context.Background(), sink, "", issue)

		require.NoError(t, err)
		require.Len(t, sink.upserted, 1)
	})

	t.Run("skips an issue that doesn't match the assignee filter", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		issue := sync.Issue{RepoFullName: "alice/widgets", Number: 5, Assignees: []string{"bob"}}

		err := sync.HandleIssueEvent(context.Background(), sink, "alice", issue)

		require.NoError(t, err)
		require.Empty(t, sink.upserted)
	})

	t.Run("upserts an issue that matches the assignee filter", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		issue := sync.Issue{RepoFullName: "alice/widgets", Number: 5, Assignees: []string{"alice"}}

		err := sync.HandleIssueEvent(context.Background(), sink, "alice", issue)

		require.NoError(t, err)
		require.Len(t, sink.upserted, 1)
	})
}
