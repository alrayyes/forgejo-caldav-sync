// Package forgejo adapts sync.IssueSource to the Forgejo API: a global
// issue search for reconciliation, and issue-webhook parsing/signature
// verification for the real-time path. It has no dependency on
// internal/caldav or internal/sync's orchestration — only on the Issue type
// those consume.
package forgejo

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alrayyes/forgejo-caldav-sync/internal/sync"
)

// defaultPageSize matches the search API's own default page size, so a
// caller that never overrides Client.PageSize gets the API's natural
// pagination rather than an arbitrary smaller one.
const defaultPageSize = 50

// Client talks to a Forgejo instance's REST API, authenticated with a
// personal access token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client

	// PageSize overrides the number of issues requested per page. Zero
	// means defaultPageSize; exported so tests can force pagination with a
	// small fixture instead of generating dozens of issues.
	PageSize int
}

// NewClient returns a Client for the Forgejo instance at baseURL.
func NewClient(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token, http: http.DefaultClient}
}

// ListIssues returns every issue visible to this client's token, across
// every repository, open and closed alike. Filtering by assignee happens in
// internal/sync, not here — this always returns the unfiltered set.
func (c *Client) ListIssues(ctx context.Context) ([]sync.Issue, error) {
	pageSize := c.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	var issues []sync.Issue
	for page := 1; ; page++ {
		batch, err := c.searchIssuesPage(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}
		for _, wi := range batch {
			issues = append(issues, wi.toIssue(wi.Repository.FullName))
		}
		if len(batch) < pageSize {
			return issues, nil
		}
	}
}

func (c *Client) searchIssuesPage(ctx context.Context, page, limit int) ([]searchResultIssue, error) {
	u := fmt.Sprintf("%s/api/v1/repos/issues/search?state=all&type=issues&limit=%d&page=%d",
		c.baseURL, limit, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("forgejo: building search request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("forgejo: searching issues (page %d): %w", page, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forgejo: searching issues (page %d): unexpected status %s", page, resp.Status)
	}

	var batch []searchResultIssue
	if err := json.NewDecoder(resp.Body).Decode(&batch); err != nil {
		return nil, fmt.Errorf("forgejo: decoding issue search response: %w", err)
	}
	return batch, nil
}

// wireAssignee is the subset of a Forgejo user object this tool reads off
// an issue's assignees.
type wireAssignee struct {
	Login string `json:"login"`
}

// wireIssueCore is the subset of a Forgejo issue shared by both the search
// API's response and a webhook's "issue" object.
type wireIssueCore struct {
	Number    int64          `json:"number"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	HTMLURL   string         `json:"html_url"`
	State     string         `json:"state"`
	DueDate   *time.Time     `json:"due_date"`
	ClosedAt  *time.Time     `json:"closed_at"`
	Assignees []wireAssignee `json:"assignees"`
}

func (wi wireIssueCore) toIssue(repoFullName string) sync.Issue {
	assignees := make([]string, len(wi.Assignees))
	for i, a := range wi.Assignees {
		assignees[i] = a.Login
	}
	return sync.Issue{
		RepoFullName: repoFullName,
		Number:       wi.Number,
		Title:        wi.Title,
		Body:         wi.Body,
		URL:          wi.HTMLURL,
		State:        wi.State,
		Assignees:    assignees,
		DueDate:      wi.DueDate,
		ClosedAt:     wi.ClosedAt,
	}
}

// searchResultIssue is one element of the issue search API's response: the
// issue itself plus which repository it belongs to.
type searchResultIssue struct {
	wireIssueCore
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// webhookPayload is a Forgejo "issues" webhook delivery.
type webhookPayload struct {
	Action     string        `json:"action"`
	Issue      wireIssueCore `json:"issue"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// ParseIssueWebhook decodes an "issues" webhook delivery body into the
// action Forgejo reported ("opened", "closed", "edited", ...) and the issue
// it applies to.
func ParseIssueWebhook(body []byte) (action string, issue sync.Issue, err error) {
	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", sync.Issue{}, fmt.Errorf("forgejo: decoding issue webhook: %w", err)
	}
	return payload.Action, payload.Issue.toIssue(payload.Repository.FullName), nil
}

// VerifySignature reports whether signature (as sent in the
// X-Forgejo-Signature/X-Gitea-Signature header) is the hex-encoded
// HMAC-SHA256 of body keyed by secret. Uses a constant-time comparison so
// timing can't leak how much of the signature matched.
func VerifySignature(secret string, body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
