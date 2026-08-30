package forgejo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/forgejo-caldav-sync/internal/forgejo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListIssuesReturnsMappedIssuesAcrossPages(t *testing.T) {
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
	// page 2 is empty — a full first page (limit reached) followed by an
	// empty page is what stops pagination.
	var requests int
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

	issues, err := client.ListIssues(context.Background())

	require.NoError(t, err)
	require.Equal(t, 2, requests)
	require.Len(t, issues, 1)
	require.Equal(t, "alice/widgets", issues[0].RepoFullName)
	require.Equal(t, int64(1), issues[0].Number)
	require.Equal(t, "First", issues[0].Title)
	require.Equal(t, "open", issues[0].State)
	require.Equal(t, []string{"alice"}, issues[0].Assignees)
}

func TestListIssuesReturnsErrorOnNonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := forgejo.NewClient(srv.URL, "bad-token")

	_, err := client.ListIssues(context.Background())

	require.Error(t, err)
}

func TestVerifySignatureAcceptsAMatchingHMAC(t *testing.T) {
	body := []byte(`{"action":"opened"}`)
	// echo -n '{"action":"opened"}' | openssl dgst -sha256 -hmac "secret"
	const validSignature = "d42142b53efbc7cf5cd20b6e074eb33707e0de3b368f698e6d6f6c824ffb8d37"

	require.True(t, forgejo.VerifySignature("secret", body, validSignature))
	require.False(t, forgejo.VerifySignature("secret", body, "0000"))
	require.False(t, forgejo.VerifySignature("wrong-secret", body, validSignature))
}

func TestParseIssueWebhookMapsPayloadToAnIssue(t *testing.T) {
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

	require.NoError(t, err)
	require.Equal(t, "closed", action)
	require.Equal(t, "alice/widgets", issue.RepoFullName)
	require.Equal(t, int64(7), issue.Number)
	require.Equal(t, "closed", issue.State)
	require.Equal(t, []string{"bob"}, issue.Assignees)
	require.NotNil(t, issue.ClosedAt)
}

func TestParseIssueWebhookRejectsInvalidJSON(t *testing.T) {
	_, _, err := forgejo.ParseIssueWebhook([]byte("not json"))

	require.Error(t, err)
}
