package forgejo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/forgejo-caldav-sync/internal/forgejo"
	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listIssuesAcrossTwoPages runs ListIssues against a server serving one
// issue on page 1 and an empty page 2 — a full first page (limit reached)
// followed by an empty page is what stops pagination.
func listIssuesAcrossTwoPages(t *testing.T) (issues []sync.Issue, requests int, err error) {
	t.Helper()

	page1 := []map[string]any{
		{
			"number":   1,
			"title":    "First",
			"body":     "body one",
			"html_url": "https://forge.example.com/alice/widgets/issues/1",
			"state":    "open",
			"assignees": []map[string]any{
				{"login": "alice"},
			},
			"repository": map[string]any{"full_name": "alice/widgets"},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assert, not require: require.FailNow from inside a handler
		// goroutine doesn't stop the test the way it would on the test's
		// own goroutine.
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		requests++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode([]map[string]any{})

			return
		}
		_ = json.NewEncoder(w).Encode(page1)
	}))
	defer srv.Close()

	client := forgejo.NewClient(srv.URL, "test-token")
	client.PageSize = 1

	issues, err = client.ListIssues(context.Background())
	if err != nil {
		return issues, requests, fmt.Errorf("test fixture: %w", err)
	}

	return issues, requests, nil
}

func TestListIssuesPagination(t *testing.T) {
	t.Parallel()

	issues, requests, err := listIssuesAcrossTwoPages(t)

	t.Run("succeeds", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, err)
	})

	t.Run("stops at the first empty page", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 2, requests)
	})

	t.Run("returns one issue", func(t *testing.T) {
		t.Parallel()
		require.Len(t, issues, 1)
	})
}

func TestListIssuesMapping(t *testing.T) {
	t.Parallel()

	issues, _, err := listIssuesAcrossTwoPages(t)
	require.NoError(t, err)
	require.Len(t, issues, 1)

	t.Run("maps the repository name", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "alice/widgets", issues[0].RepoFullName)
	})

	t.Run("maps the issue number", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, int64(1), issues[0].Number)
	})

	t.Run("maps the title", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "First", issues[0].Title)
	})

	t.Run("maps the state", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "open", issues[0].State)
	})

	t.Run("maps the assignees", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, []string{"alice"}, issues[0].Assignees)
	})
}

func TestListIssuesReturnsErrorOnNonSuccessStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := forgejo.NewClient(srv.URL, "bad-token")

	_, err := client.ListIssues(context.Background())

	require.Error(t, err)
}

func TestVerifySignatureAcceptsAMatchingHMAC(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"opened"}`)
	// echo -n '{"action":"opened"}' | openssl dgst -sha256 -hmac "secret"
	const validSignature = "d42142b53efbc7cf5cd20b6e074eb33707e0de3b368f698e6d6f6c824ffb8d37"

	require.True(t, forgejo.VerifySignature("secret", body, validSignature))
}

func TestVerifySignatureRejectsAnIncorrectSignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"opened"}`)

	require.False(t, forgejo.VerifySignature("secret", body, "0000"))
}

func TestVerifySignatureRejectsAWrongSecret(t *testing.T) {
	t.Parallel()

	body := []byte(`{"action":"opened"}`)
	const validSignature = "d42142b53efbc7cf5cd20b6e074eb33707e0de3b368f698e6d6f6c824ffb8d37"

	require.False(t, forgejo.VerifySignature("wrong-secret", body, validSignature))
}

func TestParseIssueWebhookMapsPayloadToAnIssue(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"action": "closed",
		"issue": {
			"number": 7,
			"title": "Old bug",
			"body": "",
			"html_url": "https://forge.example.com/alice/widgets/issues/7",
			"state": "closed",
			"closed_at": "2026-08-20T12:00:00Z",
			"assignees": [{"login": "bob"}]
		},
		"repository": {"full_name": "alice/widgets"}
	}`)

	action, issue, err := forgejo.ParseIssueWebhook(payload)

	t.Run("succeeds", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, err)
	})

	t.Run("maps the action", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "closed", action)
	})

	t.Run("maps the repository name", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "alice/widgets", issue.RepoFullName)
	})

	t.Run("maps the issue number", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, int64(7), issue.Number)
	})

	t.Run("maps the state", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "closed", issue.State)
	})

	t.Run("maps the assignees", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, []string{"bob"}, issue.Assignees)
	})

	t.Run("maps the closed time", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, issue.ClosedAt)
	})
}

func TestParseIssueWebhookRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	_, _, err := forgejo.ParseIssueWebhook([]byte("not json"))

	require.Error(t, err)
}
