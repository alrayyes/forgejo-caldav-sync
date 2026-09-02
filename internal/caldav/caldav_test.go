package caldav_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alrayyes/forgejo-caldav-sync/internal/caldav"
	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
	"github.com/stretchr/testify/require"
)

func dueAt(t *testing.T, s string) *time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)

	return &tm
}

// recordedUpsert runs client.Upsert(todo) against a test server and
// captures the request it sent.
type recordedUpsert struct {
	method, path, authUser, authPass, body string
	err                                    error
}

func recordUpsert(t *testing.T, todo sync.Todo) recordedUpsert {
	t.Helper()
	var rec recordedUpsert
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.authUser, rec.authPass, _ = r.BasicAuth()
		body, _ := io.ReadAll(r.Body)
		rec.body = string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "secret")
	rec.err = client.Upsert(context.Background(), todo)

	return rec
}

func TestUpsertPutsAnICSObjectAtTheUIDPath(t *testing.T) {
	t.Parallel()

	rec := recordUpsert(t, sync.Todo{
		UID:         "forgejo-caldav-sync-alice-widgets-42",
		Summary:     "alice/widgets#42: Widget falls over",
		Description: "Steps to reproduce...",
		URL:         "https://forge.example.com/alice/widgets/issues/42",
		Due:         dueAt(t, "2026-09-01T00:00:00Z"),
	})

	t.Run("succeeds", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, rec.err)
	})

	t.Run("PUTs to the UID-derived path", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, http.MethodPut, rec.method)
		require.Equal(t, "/calendars/alice/forgejo/forgejo-caldav-sync-alice-widgets-42.ics", rec.path)
	})

	t.Run("authenticates with the configured credentials", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "alice", rec.authUser)
		require.Equal(t, "secret", rec.authPass)
	})

	t.Run("body is a VTODO", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "BEGIN:VTODO")
	})

	t.Run("body carries the UID", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "UID:forgejo-caldav-sync-alice-widgets-42")
	})

	t.Run("body carries the summary", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "SUMMARY:alice/widgets#42: Widget falls over")
	})

	t.Run("body is not done", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "STATUS:NEEDS-ACTION")
	})

	t.Run("body carries the due date", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "DUE")
	})

	t.Run("body carries the URL", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "URL:https://forge.example.com/alice/widgets/issues/42")
	})
}

func TestUpsertMarksDoneIssuesCompleted(t *testing.T) {
	t.Parallel()

	rec := recordUpsert(t, sync.Todo{
		UID:       "forgejo-caldav-sync-alice-widgets-7",
		Summary:   "alice/widgets#7: Old bug",
		Done:      true,
		Completed: dueAt(t, "2026-08-20T12:00:00Z"),
	})

	t.Run("succeeds", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, rec.err)
	})

	t.Run("body status is completed", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "STATUS:COMPLETED")
	})

	t.Run("body carries a completed time", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, rec.body, "COMPLETED")
	})
}

func TestUpsertReturnsErrorOnNonSuccessStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "wrong")

	err := client.Upsert(context.Background(), sync.Todo{UID: "x", Summary: "x"})

	require.Error(t, err)
}

func TestEnsureCollectionCreatesItWithMkcalendar(t *testing.T) {
	t.Parallel()

	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "secret")

	err := client.EnsureCollection(context.Background())

	require.NoError(t, err)
	require.Equal(t, "MKCALENDAR", gotMethod)
}

func TestEnsureCollectionToleratesAnAlreadyExistingCollection(t *testing.T) {
	t.Parallel()

	statuses := map[string]int{
		"405 Method Not Allowed": http.StatusMethodNotAllowed,
		"409 Conflict":           http.StatusConflict,
	}
	for name, status := range statuses {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))
			defer srv.Close()

			client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "secret")
			err := client.EnsureCollection(context.Background())

			require.NoError(t, err)
		})
	}
}

func TestEnsureCollectionReturnsErrorOnOtherFailures(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "wrong")

	err := client.EnsureCollection(context.Background())

	require.Error(t, err)
}

func TestNewClientRequiresATrailingSlashOnBaseURL(t *testing.T) {
	t.Parallel()

	client := caldav.NewClient("https://dav.example.com/calendars/alice/forgejo", "alice", "secret")

	require.True(t, strings.HasSuffix(client.CollectionURL(), "/"))
}
