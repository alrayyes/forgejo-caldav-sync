package api_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/forgejo-caldav-sync/internal/api"
	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
	"github.com/stretchr/testify/require"
)

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

const (
	webhookSecret = "test-secret"
	testRepo      = "alice/widgets"
	testAssignee  = "bob"
)

func sign(body []byte) string {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

func issuePayload(t *testing.T) []byte {
	t.Helper()
	payload := map[string]any{
		"action": "opened",
		"issue": map[string]any{
			"number":   1,
			"title":    "Test issue",
			"body":     "body",
			"html_url": "https://forge.example.com/" + testRepo + "/issues/1",
			"state":    "open",
			"assignees": []map[string]any{
				{"login": testAssignee},
			},
		},
		"repository": map[string]any{"full_name": testRepo},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	return body
}

// postWebhook sends body to /webhooks/forgejo through sink's mux, with the
// given event type and signature headers, and returns the recorded
// response.
func postWebhook(t *testing.T, sink *fakeSink, assignee, event string, body, signature []byte) *httptest.ResponseRecorder {
	t.Helper()
	mux := api.NewMux(sink, webhookSecret, assignee)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	if event != "" {
		req.Header.Set("X-Forgejo-Event", event)
	}
	if signature != nil {
		req.Header.Set("X-Forgejo-Signature", string(signature))
	}
	mux.ServeHTTP(rec, req)

	return rec
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	mux := api.NewMux(&fakeSink{}, webhookSecret, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWebhookValidIssueEvent(t *testing.T) {
	t.Parallel()
	body := issuePayload(t)

	t.Run("responds 202 Accepted", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		rec := postWebhook(t, sink, "", "issues", body, []byte(sign(body)))

		require.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("upserts the issue", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		postWebhook(t, sink, "", "issues", body, []byte(sign(body)))

		require.Len(t, sink.upserted, 1)
	})
}

func TestWebhookMissingSignature(t *testing.T) {
	t.Parallel()
	body := issuePayload(t)

	t.Run("responds 401 Unauthorized", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		rec := postWebhook(t, sink, "", "issues", body, nil)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("does not upsert", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		postWebhook(t, sink, "", "issues", body, nil)

		require.Empty(t, sink.upserted)
	})
}

func TestWebhookIncorrectSignature(t *testing.T) {
	t.Parallel()
	body := issuePayload(t)

	t.Run("responds 401 Unauthorized", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		rec := postWebhook(t, sink, "", "issues", body, []byte("0000"))

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("does not upsert", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		postWebhook(t, sink, "", "issues", body, []byte("0000"))

		require.Empty(t, sink.upserted)
	})
}

func TestWebhookNonIssueEvent(t *testing.T) {
	t.Parallel()
	body := issuePayload(t)

	t.Run("responds 204 No Content", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		rec := postWebhook(t, sink, "", "push", body, []byte(sign(body)))

		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("does not upsert", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		postWebhook(t, sink, "", "push", body, []byte(sign(body)))

		require.Empty(t, sink.upserted)
	})
}

func TestWebhookAssigneeFilterExcludesTheIssue(t *testing.T) {
	t.Parallel()
	body := issuePayload(t)

	t.Run("responds 202 Accepted", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		rec := postWebhook(t, sink, "alice", "issues", body, []byte(sign(body)))

		require.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("does not upsert", func(t *testing.T) {
		t.Parallel()
		sink := &fakeSink{}
		postWebhook(t, sink, "alice", "issues", body, []byte(sign(body)))

		require.Empty(t, sink.upserted)
	})
}

func TestWebhookMalformedJSON(t *testing.T) {
	t.Parallel()
	body := []byte("not json")
	sink := &fakeSink{}

	rec := postWebhook(t, sink, "", "issues", body, []byte(sign(body)))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
