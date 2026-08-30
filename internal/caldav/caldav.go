// Package caldav adapts sync.CalendarSink to a real CalDAV server: an
// iCalendar VTODO, PUT to a URL derived from its UID. It doesn't
// distinguish create from update — a PUT to an existing resource overwrites
// it, which is what makes Upsert idempotent for free.
package caldav

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
	ics "github.com/arran4/golang-ical"
)

// errUnexpectedStatus is wrapped with the actual status text rather than
// building a one-off dynamic error, so callers can match on it with
// errors.Is if they ever need to.
var errUnexpectedStatus = errors.New("unexpected status")

// Client talks to one CalDAV collection (a Baïkal task list, or any other
// standards-compliant CalDAV server) over basic auth.
type Client struct {
	collectionURL string
	username      string
	password      string
	http          *http.Client
}

// NewClient returns a Client for the collection at baseURL, which is
// normalized to always end in "/" so UID-derived resource URLs join onto it
// cleanly.
func NewClient(baseURL, username, password string) *Client {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Client{
		collectionURL: baseURL,
		username:      username,
		password:      password,
		http:          http.DefaultClient,
	}
}

// CollectionURL returns the normalized collection URL this client writes
// to.
func (c *Client) CollectionURL() string {
	return c.collectionURL
}

func (c *Client) todoURL(uid string) string {
	return c.collectionURL + url.PathEscape(uid) + ".ics"
}

// EnsureCollection creates the task list this client points at if it
// doesn't already exist. Idempotent: a server reporting the collection
// already exists (405 or 409, both seen in the wild for a repeated
// MKCALENDAR) is treated as success, not an error.
func (c *Client) EnsureCollection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "MKCALENDAR", c.collectionURL, nil)
	if err != nil {
		return fmt.Errorf("caldav: building MKCALENDAR request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("caldav: creating collection: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusMethodNotAllowed, resp.StatusCode == http.StatusConflict:
		return nil // already exists
	default:
		return fmt.Errorf("caldav: creating collection: %w: %s", errUnexpectedStatus, resp.Status)
	}
}

// Upsert writes todo to the CalDAV collection, creating or overwriting the
// VTODO at the URL its UID derives.
func (c *Client) Upsert(ctx context.Context, todo sync.Todo) error {
	body := strings.NewReader(toICS(todo))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.todoURL(todo.UID), body)
	if err != nil {
		return fmt.Errorf("caldav: building PUT request: %w", err)
	}
	req.Header.Set("Content-Type", "text/calendar; charset=utf-8")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("caldav: upserting %s: %w", todo.UID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("caldav: upserting %s: %w: %s", todo.UID, errUnexpectedStatus, resp.Status)
	}

	return nil
}

func toICS(todo sync.Todo) string {
	cal := ics.NewCalendar()
	cal.SetProductId("-//forgejo-caldav-sync//EN")

	vtodo := cal.AddTodo(todo.UID)
	vtodo.SetSummary(todo.Summary)
	if todo.Description != "" {
		vtodo.SetDescription(todo.Description)
	}
	if todo.URL != "" {
		vtodo.SetURL(todo.URL)
	}
	if todo.Due != nil {
		vtodo.SetDueAt(*todo.Due)
	}

	if todo.Done {
		vtodo.SetStatus(ics.ObjectStatusCompleted)
		vtodo.SetPercentComplete(100)
		if todo.Completed != nil {
			vtodo.SetCompletedAt(*todo.Completed)
		}
	} else {
		vtodo.SetStatus(ics.ObjectStatusNeedsAction)
	}

	return cal.Serialize()
}
