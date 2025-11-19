package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
	"github.com/6ermvH/avito-reviewchecker/internal/repository"
)

type Service struct {
	repo Repository
}

var (
	ErrReviewerNotAssigned    = errors.New("reviewer not assigned to PR")
	ErrNoReplacementCandidate = errors.New("no active candidate in team")
	ErrPRMerged               = errors.New("pull request already merged")
)

type Repository interface {
	GetTeamByName(ctx context.Context, name string) (model.Team, error)
	CreateTeam(ctx context.Context, name string) (model.Team, error)
	UpsertTeamMembers(ctx context.Context, teamID string, users []model.User) error
	ListTeamMembers(ctx context.Context, teamID string) ([]model.User, error)
	GetUserByID(ctx context.Context, userID string) (model.User, error)

	SetUserActivity(ctx context.Context, userID string, active bool) (model.User, error)

	CreatePullRequest(ctx context.Context, pr model.PullRequest) (model.PullRequest, error)
	GetPullRequest(ctx context.Context, prID string) (model.PullRequest, error)
	UpdatePullRequestStatus(ctx context.Context, prID string, status model.PRStatus, mergedAt *time.Time) (model.PullRequest, error)
	ListReviewerPullRequests(ctx context.Context, userID string) ([]model.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) (model.PullRequest, error)
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateTeam(ctx context.Context, teamName string, users []model.User) error {
	team, err := s.repo.GetTeamByName(ctx, teamName)
	switch {
	case err == repository.ErrNotFound:
		team, err = s.repo.CreateTeam(ctx, teamName)
	case err != nil:
		return fmt.Errorf("find team %q: %w", teamName, err)
	}
	if err != nil {
		return fmt.Errorf("create team %q: %w", teamName, err)
	}

	if err := s.repo.UpsertTeamMembers(ctx, team.ID, users); err != nil {
		return fmt.Errorf("upsert team %q members: %w", teamName, err)
	}

	return nil
}

func (s *Service) GetTeam(ctx context.Context, teamName string) (model.Team, []model.User, error) {
	team, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		return team, nil, fmt.Errorf("find team %q: %w", teamName, err)
	}

	users, err := s.repo.ListTeamMembers(ctx, team.ID)
	if err != nil {
		return team, users, fmt.Errorf("find users from team %q: %w", teamName, err)
	}

	return team, users, nil
}

func (s *Service) ListReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	pullRequests, err := s.repo.ListReviewerPullRequests(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find prs, where user %q reviewed: %w", userID, err)
	}

	return pullRequests, nil
}

func (s *Service) SetUserActive(ctx context.Context, userID string, active bool) (model.User, error) {
	user, err := s.repo.SetUserActivity(ctx, userID, active)
	if err != nil {
		return model.User{}, fmt.Errorf("change is_active to user %q: %w", userID, err)
	}

	return user, nil
}

func (s *Service) CreatePR(ctx context.Context, prID, prName, authorID string) (model.PullRequest, error) {
	author, err := s.repo.GetUserByID(ctx, authorID)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("find author %q: %w", authorID, err)
	}

	members, err := s.repo.ListTeamMembers(ctx, author.TeamID)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("list team members for author %q: %w", authorID, err)
	}

	reviewerIDs := selectInitialReviewers(author.ID, members)

	pr := model.PullRequest{
		ID:        prID,
		Name:      prName,
		AuthorID:  author.ID,
		Status:    model.PRStatusOpen,
		Reviewers: reviewerIDs,
	}

	created, err := s.repo.CreatePullRequest(ctx, pr)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("create pr with id %q name %q for user %q: %w", prID, prName, authorID, err)
	}

	return created, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (model.PullRequest, error) {
	pr, err := s.repo.GetPullRequest(ctx, prID)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("set pr %q is merged: %w", prID, err)
	}

	if pr.Status == model.PRStatusMerged {
		return pr, nil
	}

	now := time.Now().UTC()
	pr, err = s.repo.UpdatePullRequestStatus(ctx, prID, model.PRStatusMerged, &now)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("set pr %q is merged: %w", prID, err)
	}

	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (model.PullRequest, string, error) {
	pr, err := s.repo.GetPullRequest(ctx, prID)
	if err != nil {
		return model.PullRequest{}, "", fmt.Errorf("find pr %q: %w", prID, err)
	}
	if pr.Status == model.PRStatusMerged {
		return model.PullRequest{}, "", ErrPRMerged
	}

	if !isReviewerAssigned(pr, oldUserID) {
		return model.PullRequest{}, "", ErrReviewerNotAssigned
	}

	reviewer, err := s.repo.GetUserByID(ctx, oldUserID)
	if err != nil {
		return model.PullRequest{}, "", fmt.Errorf("find reviewer %q: %w", oldUserID, err)
	}

	members, err := s.repo.ListTeamMembers(ctx, reviewer.TeamID)
	if err != nil {
		return model.PullRequest{}, "", fmt.Errorf("list team members for team %q: %w", reviewer.TeamID, err)
	}

	candidates := filterCandidates(members, pr, oldUserID)
	if len(candidates) == 0 {
		return model.PullRequest{}, "", ErrNoReplacementCandidate
	}

	targetID := candidates[0]

	updated, err := s.repo.ReplaceReviewer(ctx, prID, oldUserID, targetID)
	if err != nil {
		return model.PullRequest{}, "", fmt.Errorf("replace reviewer %q -> %q for pr %q: %w", oldUserID, targetID, prID, err)
	}

	return updated, targetID, nil
}

func isReviewerAssigned(pr model.PullRequest, reviewerID string) bool {
	for _, id := range pr.Reviewers {
		if id == reviewerID {
			return true
		}
	}
	return false
}

func filterCandidates(members []model.User, pr model.PullRequest, removedReviewer string) []string {
	existing := make(map[string]struct{}, len(pr.Reviewers))
	for _, reviewer := range pr.Reviewers {
		existing[reviewer] = struct{}{}
	}

	candidates := make([]string, 0)
	for _, member := range members {
		if !member.IsActive {
			continue
		}
		if member.ID == pr.AuthorID {
			continue
		}
		if member.ID == removedReviewer {
			continue
		}
		if _, ok := existing[member.ID]; ok {
			continue
		}
		candidates = append(candidates, member.ID)
	}
	return candidates
}

func selectInitialReviewers(authorID string, members []model.User) []string {
	candidates := make([]string, 0, len(members))
	for _, member := range members {
		if !member.IsActive {
			continue
		}
		if member.ID == authorID {
			continue
		}
		candidates = append(candidates, member.ID)
	}

	if len(candidates) > 2 {
		return candidates[:2]
	}
	return candidates
}
