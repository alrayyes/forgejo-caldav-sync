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

func TestUpsertPutsAnICSObjectAtTheUIDPath(t *testing.T) {
	var gotMethod, gotPath, gotAuthUser, gotAuthPass, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuthUser, gotAuthPass, _ = r.BasicAuth()
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "secret")
	todo := sync.Todo{
		UID:         "forgejo-caldav-sync-alice-widgets-42",
		Summary:     "alice/widgets#42: Widget falls over",
		Description: "Steps to reproduce...",
		URL:         "https://forge.example.com/alice/widgets/issues/42",
		Due:         dueAt(t, "2026-09-01T00:00:00Z"),
	}

	err := client.Upsert(context.Background(), todo)

	require.NoError(t, err)
	require.Equal(t, http.MethodPut, gotMethod)
	require.Equal(t, "/calendars/alice/forgejo/forgejo-caldav-sync-alice-widgets-42.ics", gotPath)
	require.Equal(t, "alice", gotAuthUser)
	require.Equal(t, "secret", gotAuthPass)
	require.Contains(t, gotBody, "BEGIN:VTODO")
	require.Contains(t, gotBody, "UID:forgejo-caldav-sync-alice-widgets-42")
	require.Contains(t, gotBody, "SUMMARY:alice/widgets#42: Widget falls over")
	require.Contains(t, gotBody, "STATUS:NEEDS-ACTION")
	require.Contains(t, gotBody, "DUE")
	require.Contains(t, gotBody, "URL:https://forge.example.com/alice/widgets/issues/42")
}

func TestUpsertMarksDoneIssuesCompleted(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "secret")
	todo := sync.Todo{
		UID:       "forgejo-caldav-sync-alice-widgets-7",
		Summary:   "alice/widgets#7: Old bug",
		Done:      true,
		Completed: dueAt(t, "2026-08-20T12:00:00Z"),
	}

	err := client.Upsert(context.Background(), todo)

	require.NoError(t, err)
	require.Contains(t, gotBody, "STATUS:COMPLETED")
	require.Contains(t, gotBody, "COMPLETED")
}

func TestUpsertReturnsErrorOnNonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "wrong")

	err := client.Upsert(context.Background(), sync.Todo{UID: "x", Summary: "x"})

	require.Error(t, err)
}

func TestEnsureCollectionCreatesItWithMkcalendar(t *testing.T) {
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
	for _, status := range []int{http.StatusMethodNotAllowed, http.StatusConflict} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		}))

		client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "secret")
		err := client.EnsureCollection(context.Background())

		require.NoErrorf(t, err, "status %d should be treated as already-exists", status)
		srv.Close()
	}
}

func TestEnsureCollectionReturnsErrorOnOtherFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := caldav.NewClient(srv.URL+"/calendars/alice/forgejo/", "alice", "wrong")

	err := client.EnsureCollection(context.Background())

	require.Error(t, err)
}

func TestNewClientRequiresATrailingSlashOnBaseURL(t *testing.T) {
	client := caldav.NewClient("https://dav.example.com/calendars/alice/forgejo", "alice", "secret")

	require.True(t, strings.HasSuffix(client.CollectionURL(), "/"))
}
