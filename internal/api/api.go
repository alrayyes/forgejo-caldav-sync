// Package api is the HTTP delivery layer: liveness, and the Forgejo issue
// webhook that drives real-time sync. It translates the wire format
// (headers, JSON body) into calls against internal/sync's domain types —
// the mapping and filtering logic itself lives there, not here.
package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alrayyes/forgejo-caldav-sync/internal/forgejo"
	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
)

// signatureHeaders are checked in order — Forgejo sends its own header, but
// also X-Gitea-Signature for compatibility with tooling written against
// Gitea's webhook format, so either satisfies verification.
var signatureHeaders = []string{"X-Forgejo-Signature", "X-Gitea-Signature"}

// NewMux wires the health and webhook handlers. sink is where a matching
// issue event gets upserted; webhookSecret verifies inbound deliveries;
// assignee restricts sync the same way it does for reconciliation.
func NewMux(sink sync.CalendarSink, webhookSecret, assignee string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("POST /webhooks/forgejo", handleWebhook(sink, webhookSecret, assignee))

	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleWebhook(sink sync.CalendarSink, webhookSecret, assignee string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "reading body", http.StatusBadRequest)

			return
		}

		if !hasValidSignature(webhookSecret, body, r.Header) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)

			return
		}

		event := firstHeader(r.Header, "X-Forgejo-Event", "X-Gitea-Event")
		if event != "issues" {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		_, issue, err := forgejo.ParseIssueWebhook(body)
		if err != nil {
			http.Error(w, "malformed payload", http.StatusBadRequest)

			return
		}

		if err := sync.HandleIssueEvent(r.Context(), sink, assignee, issue); err != nil {
			http.Error(w, "sync failed", http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func hasValidSignature(secret string, body []byte, header http.Header) bool {
	for _, name := range signatureHeaders {
		if sig := header.Get(name); sig != "" {
			return forgejo.VerifySignature(secret, body, sig)
		}
	}

	return false
}

func firstHeader(header http.Header, names ...string) string {
	for _, name := range names {
		if v := header.Get(name); v != "" {
			return v
		}
	}

	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
