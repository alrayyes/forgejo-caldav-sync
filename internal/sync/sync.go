// Package sync holds the mapping and orchestration logic that turns Forgejo
// issues into CalDAV VTODOs. It knows nothing about HTTP, the Forgejo API,
// or CalDAV wire formats — those live in internal/forgejo and
// internal/caldav, which adapt to the two narrow interfaces defined here.
package sync

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Issue is the subset of a Forgejo issue this tool cares about, independent
// of whether it came from a webhook payload or a search API response.
type Issue struct {
	RepoFullName string // "owner/repo"
	Number       int64
	Title        string
	Body         string
	URL          string
	State        string // "open" or "closed"
	Assignees    []string
	DueDate      *time.Time
	ClosedAt     *time.Time
}

// Todo is a CalDAV VTODO, independent of how it gets serialized to iCalendar
// or transported over CalDAV.
type Todo struct {
	UID         string
	Summary     string
	Description string
	URL         string
	Due         *time.Time
	Completed   *time.Time
	Done        bool
}

// IssueSource is the one Forgejo capability this package needs: every issue
// visible to whatever token the caller configured, unfiltered — filtering by
// assignee happens here, not in the adapter, so it stays testable without a
// real Forgejo instance.
type IssueSource interface {
	ListIssues(ctx context.Context) ([]Issue, error)
}

// CalendarSink is the one CalDAV capability this package needs: create the
// VTODO if it doesn't exist, or overwrite it in place if it does. UID is
// what makes that idempotent.
type CalendarSink interface {
	Upsert(ctx context.Context, todo Todo) error
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// TodoUID deterministically derives a VTODO's UID from the issue it
// represents, so re-syncing the same issue always resolves to the same
// CalDAV resource instead of creating a duplicate.
func TodoUID(repoFullName string, number int64) string {
	slug := nonAlnum.ReplaceAllString(strings.ToLower(repoFullName), "-")
	slug = strings.Trim(slug, "-")

	return fmt.Sprintf("forgejo-caldav-sync-%s-%d", slug, number)
}

// ToTodo maps a Forgejo issue onto the VTODO that represents it. Pure and
// deterministic: the same issue always produces the same Todo, which is
// what makes Upsert safe to call on every sync pass rather than only once.
func ToTodo(issue Issue) Todo {
	todo := Todo{
		UID:     TodoUID(issue.RepoFullName, issue.Number),
		Summary: fmt.Sprintf("%s#%d: %s", issue.RepoFullName, issue.Number, issue.Title),
		Description: strings.TrimSpace(fmt.Sprintf("%s\n\n%s",
			issue.Body, issue.URL)),
		URL:  issue.URL,
		Due:  issue.DueDate,
		Done: issue.State == "closed",
	}
	if todo.Done {
		todo.Completed = issue.ClosedAt
	}

	return todo
}

// FilterByAssignee returns only the issues assigned to username, matched
// case-insensitively since Forgejo usernames aren't case sensitive. An
// empty username is "no filter" — every issue passes, which is what makes
// this tool sync everything by default and only narrow down when a caller
// asks for it.
func FilterByAssignee(issues []Issue, username string) []Issue {
	if username == "" {
		return issues
	}
	var filtered []Issue
	for _, issue := range issues {
		if hasAssignee(issue, username) {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

func hasAssignee(issue Issue, username string) bool {
	for _, a := range issue.Assignees {
		if strings.EqualFold(a, username) {
			return true
		}
	}

	return false
}

// Reconcile lists every issue IssueSource can see, keeps the ones matching
// assignee (all of them, if assignee is empty), and upserts each into sink.
// It's what backfills older issues on first run and self-heals anything a
// missed webhook left out of sync on every later run — there's no separate
// one-off import path.
func Reconcile(ctx context.Context, src IssueSource, sink CalendarSink, assignee string) (int, error) {
	issues, err := src.ListIssues(ctx)
	if err != nil {
		return 0, fmt.Errorf("sync: listing issues: %w", err)
	}

	matched := FilterByAssignee(issues, assignee)
	for _, issue := range matched {
		if err := sink.Upsert(ctx, ToTodo(issue)); err != nil {
			return 0, fmt.Errorf("sync: upserting issue %s#%d: %w", issue.RepoFullName, issue.Number, err)
		}
	}

	return len(matched), nil
}

// HandleIssueEvent is the webhook path: one issue, upserted immediately if
// it matches assignee, skipped without error otherwise. Skipping rather
// than deleting an existing VTODO on a non-matching re-assignment is a
// known, documented limitation — see the README.
func HandleIssueEvent(ctx context.Context, sink CalendarSink, assignee string, issue Issue) error {
	if assignee != "" && !hasAssignee(issue, assignee) {
		return nil
	}
	if err := sink.Upsert(ctx, ToTodo(issue)); err != nil {
		return fmt.Errorf("sync: upserting issue %s#%d: %w", issue.RepoFullName, issue.Number, err)
	}

	return nil
}
