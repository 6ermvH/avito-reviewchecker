package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	httpmodel "github.com/6ermvH/avito-reviewchecker/internal/model/http"
)

const (
	defaultBaseURL      = "http://app:8080"
	serviceReadyTimeout = 30 * time.Second
	servicePollInterval = time.Second
	requestTimeout      = 10 * time.Second
)

func TestEndToEndFlow(t *testing.T) {

	client := &http.Client{Timeout: requestTimeout}
	waitForServiceReady(t, client)

	teamName := fmt.Sprintf("e2e-team-%d", time.Now().UnixNano())
	members := []httpmodel.TeamMember{
		{UserID: teamName + "-author", Username: "Author", IsActive: true},
		{UserID: teamName + "-rev1", Username: "Reviewer-1", IsActive: true},
		{UserID: teamName + "-rev2", Username: "Reviewer-2", IsActive: true},
		{UserID: teamName + "-rev3", Username: "Reviewer-3", IsActive: true},
	}

	addTeam(t, client, teamName, members)

	prID := teamName + "-pr"
	pr := createPR(t, client, httpmodel.PullRequestCreateRequest{
		ID:       prID,
		Name:     "Implement new feature",
		AuthorID: members[0].UserID,
	})

	if len(pr.AssignedReviewers) < 2 {
		t.Fatalf("expected reviewers for PR %s", pr.ID)
	}

	oldReviewer := pr.AssignedReviewers[0]
	reassignResp := reassignReviewer(t, client, httpmodel.PullRequestReassignRequest{
		ID:        pr.ID,
		OldUserID: oldReviewer,
	})

	if reassignResp.ReplacedBy == "" || reassignResp.ReplacedBy == oldReviewer {
		t.Fatalf("expected reviewer replacement, got %+v", reassignResp)
	}

	if reassignResp.PR.Status != "OPEN" {
		t.Fatalf("expected PR to remain OPEN after reassign, got %s", reassignResp.PR.Status)
	}

	merged := mergePR(t, client, httpmodel.PullRequestMergeRequest{ID: pr.ID})
	if merged.Status != "MERGED" {
		t.Fatalf("expected PR to be MERGED, got %s", merged.Status)
	}

	userReviews := getUserReviews(t, client, reassignResp.ReplacedBy)
	if len(userReviews.PullRequests) == 0 {
		t.Fatalf("expected reviewer %s to have at least one PR", reassignResp.ReplacedBy)
	}

	reviewerStats := getReviewerStats(t, client)
	if len(reviewerStats.Reviewers) == 0 {
		t.Fatal("expected reviewer stats to return at least one entry")
	}
	if !containsReviewer(reviewerStats.Reviewers, "alpha-2") {
		t.Fatal("expected reviewer stats to include seeded user alpha-2")
	}

	prStats := getPRStats(t, client)
	if prStats.Total == 0 || prStats.Open+prStats.Merged == 0 {
		t.Fatalf("unexpected pull request stats: %+v", prStats)
	}
}

func addTeam(t *testing.T, client *http.Client, teamName string, members []httpmodel.TeamMember) {
	t.Helper()

	req := httpmodel.Team{
		TeamName: teamName,
		Members:  members,
	}

	var resp httpmodel.TeamResponse
	doJSONRequest(t, client, http.MethodPost, "/team/add", req, http.StatusCreated, &resp)

	if resp.Team.TeamName != teamName {
		t.Fatalf("unexpected team name %s", resp.Team.TeamName)
	}
}

func createPR(
	t *testing.T,
	client *http.Client,
	req httpmodel.PullRequestCreateRequest,
) httpmodel.PullRequest {
	t.Helper()

	var resp httpmodel.PullRequestResponse
	doJSONRequest(t, client, http.MethodPost, "/pullRequest/create", req, http.StatusCreated, &resp)

	if resp.PR.ID != req.ID {
		t.Fatalf("expected PR %s, got %s", req.ID, resp.PR.ID)
	}

	return resp.PR
}

func reassignReviewer(
	t *testing.T,
	client *http.Client,
	req httpmodel.PullRequestReassignRequest,
) httpmodel.PullRequestResponse {
	t.Helper()

	var resp httpmodel.PullRequestResponse
	doJSONRequest(t, client, http.MethodPost, "/pullRequest/reassign", req, http.StatusOK, &resp)

	return resp
}

func mergePR(
	t *testing.T,
	client *http.Client,
	req httpmodel.PullRequestMergeRequest,
) httpmodel.PullRequest {
	t.Helper()

	var resp httpmodel.PullRequestResponse
	doJSONRequest(t, client, http.MethodPost, "/pullRequest/merge", req, http.StatusOK, &resp)

	return resp.PR
}

func getUserReviews(t *testing.T, client *http.Client, userID string) httpmodel.UserReviewsResponse {
	t.Helper()

	var resp httpmodel.UserReviewsResponse
	doJSONRequest(t, client, http.MethodGet, "/users/getReview?user_id="+userID, nil, http.StatusOK, &resp)

	return resp
}

func getReviewerStats(t *testing.T, client *http.Client) httpmodel.ReviewerStatsResponse {
	t.Helper()

	var resp httpmodel.ReviewerStatsResponse
	doJSONRequest(t, client, http.MethodGet, "/stats/reviewers", nil, http.StatusOK, &resp)

	return resp
}

func getPRStats(t *testing.T, client *http.Client) httpmodel.PullRequestStatsResponse {
	t.Helper()

	var resp httpmodel.PullRequestStatsResponse
	doJSONRequest(t, client, http.MethodGet, "/stats/pullRequests", nil, http.StatusOK, &resp)

	return resp
}

func doJSONRequest(
	t *testing.T,
	client *http.Client,
	method,
	path string,
	payload interface{},
	expectedStatus int,
	out interface{},
) {
	t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL()+path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("unexpected status %d (%s): %s", resp.StatusCode, path, string(data))
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("decode response: %v\nbody: %s", err, string(data))
		}
	}
}

func waitForServiceReady(t *testing.T, client *http.Client) {
	t.Helper()

	deadline := time.Now().Add(serviceReadyTimeout)
	var lastErr error

	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, baseURL()+"/healthz", nil)
		if err != nil {
			t.Fatalf("build health request: %v", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
			lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		time.Sleep(servicePollInterval)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("timeout %s exceeded", serviceReadyTimeout)
	}

	t.Fatalf("service is not ready: %v", lastErr)
}

func baseURL() string {
	if url := os.Getenv("E2E_BASE_URL"); url != "" {
		return strings.TrimSuffix(url, "/")
	}

	return defaultBaseURL
}

func containsReviewer(stats []httpmodel.ReviewerStat, userID string) bool {
	for _, stat := range stats {
		if stat.UserID == userID {
			return true
		}
	}

	return false
}
