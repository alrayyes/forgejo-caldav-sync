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

const webhookSecret = "test-secret"

func sign(body []byte) string {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func issuePayload(t *testing.T, action, repo, assignee string) []byte {
	t.Helper()
	payload := map[string]any{
		"action": action,
		"issue": map[string]any{
			"number":   1,
			"title":    "Test issue",
			"body":     "body",
			"html_url": "https://forge.example.com/" + repo + "/issues/1",
			"state":    "open",
			"assignees": []map[string]any{
				{"login": assignee},
			},
		},
		"repository": map[string]any{"full_name": repo},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	return body
}

func TestHealthzAnswersOK(t *testing.T) {
	mux := api.NewMux(&fakeSink{}, webhookSecret, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWebhookUpsertsOnValidIssueEvent(t *testing.T) {
	sink := &fakeSink{}
	mux := api.NewMux(sink, webhookSecret, "")
	body := issuePayload(t, "opened", "alice/widgets", "bob")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	req.Header.Set("X-Forgejo-Event", "issues")
	req.Header.Set("X-Forgejo-Signature", sign(body))
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Len(t, sink.upserted, 1)
}

func TestWebhookRejectsAMissingSignature(t *testing.T) {
	sink := &fakeSink{}
	mux := api.NewMux(sink, webhookSecret, "")
	body := issuePayload(t, "opened", "alice/widgets", "bob")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	req.Header.Set("X-Forgejo-Event", "issues")
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Empty(t, sink.upserted)
}

func TestWebhookRejectsAnIncorrectSignature(t *testing.T) {
	sink := &fakeSink{}
	mux := api.NewMux(sink, webhookSecret, "")
	body := issuePayload(t, "opened", "alice/widgets", "bob")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	req.Header.Set("X-Forgejo-Event", "issues")
	req.Header.Set("X-Forgejo-Signature", "0000")
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Empty(t, sink.upserted)
}

func TestWebhookIgnoresNonIssueEvents(t *testing.T) {
	sink := &fakeSink{}
	mux := api.NewMux(sink, webhookSecret, "")
	body := issuePayload(t, "opened", "alice/widgets", "bob")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	req.Header.Set("X-Forgejo-Event", "push")
	req.Header.Set("X-Forgejo-Signature", sign(body))
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Empty(t, sink.upserted)
}

func TestWebhookSkipsIssuesNotMatchingTheAssigneeFilter(t *testing.T) {
	sink := &fakeSink{}
	mux := api.NewMux(sink, webhookSecret, "alice")
	body := issuePayload(t, "opened", "alice/widgets", "bob")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	req.Header.Set("X-Forgejo-Event", "issues")
	req.Header.Set("X-Forgejo-Signature", sign(body))
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Empty(t, sink.upserted)
}

func TestWebhookRejectsMalformedJSON(t *testing.T) {
	sink := &fakeSink{}
	mux := api.NewMux(sink, webhookSecret, "")
	body := []byte("not json")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", bytes.NewReader(body))
	req.Header.Set("X-Forgejo-Event", "issues")
	req.Header.Set("X-Forgejo-Signature", sign(body))
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
