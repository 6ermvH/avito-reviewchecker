package httpmodel

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamResponse struct {
	Team Team `json:"team"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveResponse struct {
	User User `json:"user"`
}

type UserReviewsResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type PullRequest struct {
	ID                string   `json:"pull_request_id"`
	Name              string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         string   `json:"createdAt,omitempty"`
	MergedAt          string   `json:"mergedAt,omitempty"`
}

type ErrorCode string

const (
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodePRMerged     ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned  ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate  ErrorCode = "NO_CANDIDATE"
	ErrorCodeTeamExists   ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists     ErrorCode = "PR_EXISTS"
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeInvalidInput ErrorCode = "INVALID_REQUEST"
)

type PullRequestResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by,omitempty"`
}

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

type PullRequestCreateRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type PullRequestMergeRequest struct {
	ID string `json:"pull_request_id"`
}

type PullRequestReassignRequest struct {
	ID        string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}
