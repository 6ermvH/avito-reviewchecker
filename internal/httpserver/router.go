package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
	httpmodel "github.com/6ermvH/avito-reviewchecker/internal/model/http"
	"github.com/6ermvH/avito-reviewchecker/internal/repository"
	"github.com/6ermvH/avito-reviewchecker/internal/usecase"
)

type Service interface {
	UpdateTeam(ctx context.Context, teamName string, users []model.User) error
	GetTeam(ctx context.Context, teamName string) (model.Team, []model.User, error)
	ListReviews(ctx context.Context, userID string) ([]model.PullRequest, error)
	SetUserActive(ctx context.Context, userID string, active bool) (model.User, error)
	CreatePR(ctx context.Context, prID, prName, authorID string) (model.PullRequest, error)
	MergePR(ctx context.Context, prID string) (model.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (model.PullRequest, string, error)
	ListReviewerStats(ctx context.Context) ([]model.ReviewerStat, error)
	GetPullRequestStats(ctx context.Context) (model.PullRequestStats, error)
}

func HandleTeamAdd(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req httpmodel.Team
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"invalid JSON",
			)

			return
		}

		if req.TeamName == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"team_name is required",
			)

			return
		}

		users, err := buildTeamUsers(req.Members)
		if err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				err.Error(),
			)

			return
		}

		if err := svc.UpdateTeam(r.Context(), req.TeamName, users); err != nil {
			writeDomainError(w, err, map[string]int{})

			return
		}

		team, members, err := svc.GetTeam(r.Context(), req.TeamName)
		if err != nil {
			writeError(
				w,
				http.StatusInternalServerError,
				string(httpmodel.ErrorCodeInternal),
				err.Error(),
			)

			return
		}

		writeJSON(w, http.StatusCreated, httpmodel.TeamResponse{
			Team: mapTeamResponse(team, members),
		})
	}
}

func HandleTeamGet(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"team_name is required",
			)

			return
		}

		team, members, err := svc.GetTeam(r.Context(), teamName)
		if err != nil {
			writeDomainError(w, err, map[string]int{})

			return
		}

		writeJSON(w, http.StatusOK, mapTeamResponse(team, members))
	}
}

func HandleSetUserActive(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req httpmodel.SetUserActiveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"invalid JSON",
			)

			return
		}

		if req.UserID == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"user_id is required",
			)

			return
		}

		user, err := svc.SetUserActive(r.Context(), req.UserID, req.IsActive)
		if err != nil {
			writeDomainError(w, err, map[string]int{
				repository.ErrNotFound.Error(): http.StatusNotFound,
			})

			return
		}

		writeJSON(w, http.StatusOK, httpmodel.SetUserActiveResponse{
			User: mapUserResponse(user),
		})
	}
}

func HandleGetUserReview(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"user_id is required",
			)

			return
		}

		prs, err := svc.ListReviews(r.Context(), userID)
		if err != nil {
			writeDomainError(w, err, map[string]int{})

			return
		}

		writeJSON(w, http.StatusOK, httpmodel.UserReviewsResponse{
			UserID:       userID,
			PullRequests: mapPRShortList(prs),
		})
	}
}

func HandleCreatePR(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req httpmodel.PullRequestCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"invalid JSON",
			)

			return
		}

		if req.ID == "" || req.Name == "" || req.AuthorID == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"missing required fields",
			)

			return
		}

		pr, err := svc.CreatePR(r.Context(), req.ID, req.Name, req.AuthorID)
		if err != nil {
			writeDomainError(w, err, map[string]int{})

			return
		}

		writeJSON(w, http.StatusCreated, httpmodel.PullRequestResponse{
			PR: mapPRResponse(pr),
		})
	}
}

func HandleMergePR(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req httpmodel.PullRequestMergeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"invalid JSON",
			)

			return
		}

		if req.ID == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"pull_request_id is required",
			)

			return
		}

		pr, err := svc.MergePR(r.Context(), req.ID)
		if err != nil {
			writeDomainError(w, err, map[string]int{
				repository.ErrNotFound.Error(): http.StatusNotFound,
				usecase.ErrPRMerged.Error():    http.StatusConflict,
			})

			return
		}

		writeJSON(w, http.StatusOK, httpmodel.PullRequestResponse{
			PR: mapPRResponse(pr),
		})
	}
}

func HandleReassignPR(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req httpmodel.PullRequestReassignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"invalid JSON",
			)

			return
		}

		if req.ID == "" || req.OldUserID == "" {
			writeError(
				w,
				http.StatusBadRequest,
				string(httpmodel.ErrorCodeInvalidInput),
				"pull_request_id and old_user_id are required",
			)

			return
		}

		pr, newReviewer, err := svc.ReassignReviewer(r.Context(), req.ID, req.OldUserID)
		if err != nil {
			writeDomainError(w, err, map[string]int{
				repository.ErrNotFound.Error():            http.StatusNotFound,
				usecase.ErrPRMerged.Error():               http.StatusConflict,
				usecase.ErrReviewerNotAssigned.Error():    http.StatusConflict,
				usecase.ErrNoReplacementCandidate.Error(): http.StatusConflict,
			})

			return
		}

		writeJSON(w, http.StatusOK, httpmodel.PullRequestResponse{
			PR:         mapPRResponse(pr),
			ReplacedBy: newReviewer,
		})
	}
}

func writeDomainError(w http.ResponseWriter, err error, overrides map[string]int) {
	status := http.StatusInternalServerError
	code := httpmodel.ErrorCodeInternal

	switch {
	case errors.Is(err, repository.ErrNotFound):
		status = http.StatusNotFound
		code = httpmodel.ErrorCodeNotFound
	case errors.Is(err, usecase.ErrPRMerged):
		status = http.StatusConflict
		code = httpmodel.ErrorCodePRMerged
	case errors.Is(err, usecase.ErrReviewerNotAssigned):
		status = http.StatusConflict
		code = httpmodel.ErrorCodeNotAssigned
	case errors.Is(err, usecase.ErrNoReplacementCandidate):
		status = http.StatusConflict
		code = httpmodel.ErrorCodeNoCandidate
	case errors.Is(err, usecase.ErrTeamExists):
		status = http.StatusBadRequest
		code = httpmodel.ErrorCodeTeamExists
	case errors.Is(err, usecase.ErrPullRequestExists):
		status = http.StatusConflict
		code = httpmodel.ErrorCodePRExists
	}

	if customStatus, ok := overrides[err.Error()]; ok {
		status = customStatus
	}

	writeError(w, status, string(code), err.Error())
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errchkjson
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func mapTeamResponse(team model.Team, members []model.User) httpmodel.Team {
	payload := httpmodel.Team{
		TeamName: team.Name,
		Members:  make([]httpmodel.TeamMember, 0, len(members)),
	}
	for _, member := range members {
		payload.Members = append(payload.Members, httpmodel.TeamMember{
			UserID:   member.ID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	return payload
}

func mapUserResponse(user model.User) httpmodel.User {
	return httpmodel.User{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func mapPRResponse(pr model.PullRequest) httpmodel.PullRequest {
	payload := httpmodel.PullRequest{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: append([]string(nil), pr.Reviewers...),
	}
	if !pr.CreatedAt.IsZero() {
		payload.CreatedAt = pr.CreatedAt.UTC().Format(time.RFC3339)
	}

	if pr.MergedAt != nil {
		payload.MergedAt = pr.MergedAt.UTC().Format(time.RFC3339)
	}

	return payload
}

func mapPRShortList(prs []model.PullRequest) []httpmodel.PullRequestShort {
	resp := make([]httpmodel.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		resp = append(resp, httpmodel.PullRequestShort{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   string(pr.Status),
		})
	}

	return resp
}

var (
	errUserIDRequired     = errors.New("user_id is required")
	errUsernameRequired   = errors.New("username is required")
	errDuplicateMemberIDs = errors.New("duplicate user_id in members")
)

func buildTeamUsers(members []httpmodel.TeamMember) ([]model.User, error) {
	users := make([]model.User, 0, len(members))
	seen := make(map[string]struct{}, len(members))

	for _, member := range members {
		switch {
		case member.UserID == "":
			return nil, errUserIDRequired
		case member.Username == "":
			return nil, errUsernameRequired
		}

		if _, exists := seen[member.UserID]; exists {
			return nil, errDuplicateMemberIDs
		}

		seen[member.UserID] = struct{}{}

		users = append(users, model.User{
			ID:       member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	return users, nil
}
