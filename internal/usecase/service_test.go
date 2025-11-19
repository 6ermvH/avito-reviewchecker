package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
	"github.com/6ermvH/avito-reviewchecker/internal/repository"
	mocks_repository "github.com/6ermvH/avito-reviewchecker/internal/repository/mocks"
)

func TestUpdateTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)
	members := []model.User{}

	t.Run("Good: team exists", func(t *testing.T) {
		team := model.Team{ID: uuid.NewString(), Name: "team"}
		repo.EXPECT().
			GetTeamByName(gomock.Any(), "team").
			Return(team, nil)
		repo.EXPECT().
			InsertTeamMembers(gomock.Any(), team.ID, members).
			Return(nil)

		require.NoError(t, service.UpdateTeam(context.Background(), "team", members))
	})

	t.Run("Good: create new team", func(t *testing.T) {
		team := model.Team{ID: uuid.NewString(), Name: "created"}
		repo.EXPECT().
			GetTeamByName(gomock.Any(), "created").
			Return(model.Team{}, repository.ErrNotFound)
		repo.EXPECT().
			CreateTeam(gomock.Any(), "created").
			Return(team, nil)
		repo.EXPECT().
			InsertTeamMembers(gomock.Any(), team.ID, members).
			Return(nil)

		require.NoError(t, service.UpdateTeam(context.Background(), "created", members))
	})

	t.Run("Bad: get team error", func(t *testing.T) {
		repo.EXPECT().
			GetTeamByName(gomock.Any(), "boom").
			Return(model.Team{}, errors.New("get error"))

		require.Error(t, service.UpdateTeam(context.Background(), "boom", members))
	})

	t.Run("Bad: insert members error", func(t *testing.T) {
		team := model.Team{ID: uuid.NewString(), Name: "team"}
		repo.EXPECT().
			GetTeamByName(gomock.Any(), "team").
			Return(team, nil)
		repo.EXPECT().
			InsertTeamMembers(gomock.Any(), team.ID, members).
			Return(errors.New("insert error"))

		require.Error(t, service.UpdateTeam(context.Background(), "team", members))
	})
}

func TestGetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)

	t.Run("Good: success", func(t *testing.T) {
		team := model.Team{ID: uuid.NewString(), Name: "team"}
		users := []model.User{
			{ID: uuid.NewString(), Username: "user", IsActive: true},
		}

		repo.EXPECT().
			GetTeamByName(gomock.Any(), team.Name).
			Return(team, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), team.ID).
			Return(users, nil)

		gotTeam, gotUsers, err := service.GetTeam(context.Background(), team.Name)
		require.NoError(t, err)
		require.Equal(t, team, gotTeam)
		require.Equal(t, users, gotUsers)
	})

	t.Run("Bad: no team", func(t *testing.T) {
		repo.EXPECT().
			GetTeamByName(gomock.Any(), t.Name()).
			Return(model.Team{}, repository.ErrNotFound)

		_, _, err := service.GetTeam(context.Background(), t.Name())
		require.Error(t, err)
	})

	t.Run("Bad: list members error", func(t *testing.T) {
		team := model.Team{ID: uuid.NewString(), Name: t.Name()}
		repo.EXPECT().
			GetTeamByName(gomock.Any(), team.Name).
			Return(team, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), team.ID).
			Return(nil, errors.New("list error"))

		_, _, err := service.GetTeam(context.Background(), team.Name)
		require.Error(t, err)
	})
}

func TestListReviews(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)

	t.Run("Good: success", func(t *testing.T) {
		expected := []model.PullRequest{
			{ID: "pr1"},
		}
		repo.EXPECT().
			ListReviewerPullRequests(gomock.Any(), "user").
			Return(expected, nil)

		result, err := service.ListReviews(context.Background(), "user")
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("Bad: repo error", func(t *testing.T) {
		repo.EXPECT().
			ListReviewerPullRequests(gomock.Any(), "user").
			Return(nil, errors.New("list error"))

		_, err := service.ListReviews(context.Background(), "user")
		require.Error(t, err)
	})
}

func TestSetUserActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)

	t.Run("Good: updated", func(t *testing.T) {
		user := model.User{ID: "user", IsActive: true}
		repo.EXPECT().
			SetUserActivity(gomock.Any(), "user", true).
			Return(user, nil)

		result, err := service.SetUserActive(context.Background(), "user", true)
		require.NoError(t, err)
		require.Equal(t, user, result)
	})

	t.Run("Bad: repo error", func(t *testing.T) {
		repo.EXPECT().
			SetUserActivity(gomock.Any(), "user", false).
			Return(model.User{}, errors.New("update error"))

		_, err := service.SetUserActive(context.Background(), "user", false)
		require.Error(t, err)
	})
}

func TestCreatePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)

	author := model.User{ID: "author", TeamID: "team"}
	members := []model.User{
		{ID: "author", TeamID: "team", IsActive: true},
		{ID: "u1", TeamID: "team", IsActive: true},
		{ID: "u2", TeamID: "team", IsActive: true},
		{ID: "inactive", TeamID: "team", IsActive: false},
	}

	t.Run("Good: created", func(t *testing.T) {
		repo.EXPECT().
			GetUserByID(gomock.Any(), "author").
			Return(author, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(members, nil)
		repo.EXPECT().
			CreatePullRequest(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, pr model.PullRequest) (model.PullRequest, error) {
				require.Equal(t, "pr-1", pr.ID)
				require.Equal(t, "new feature", pr.Name)
				require.Equal(t, "author", pr.AuthorID)
				require.Equal(t, model.PRStatusOpen, pr.Status)
				require.Equal(t, []string{"u1", "u2"}, pr.Reviewers)
				return pr, nil
			})

		result, err := service.CreatePR(context.Background(), "pr-1", "new feature", "author")
		require.NoError(t, err)
		require.Equal(t, "pr-1", result.ID)
	})

	t.Run("Bad: author not found", func(t *testing.T) {
		repo.EXPECT().
			GetUserByID(gomock.Any(), "missing").
			Return(model.User{}, errors.New("not found"))

		_, err := service.CreatePR(context.Background(), "pr", "name", "missing")
		require.Error(t, err)
	})

	t.Run("Bad: list team members error", func(t *testing.T) {
		repo.EXPECT().
			GetUserByID(gomock.Any(), "author").
			Return(author, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(nil, errors.New("list error"))

		_, err := service.CreatePR(context.Background(), "pr", "name", "author")
		require.Error(t, err)
	})

	t.Run("Bad: create pr error", func(t *testing.T) {
		repo.EXPECT().
			GetUserByID(gomock.Any(), "author").
			Return(author, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(members, nil)
		repo.EXPECT().
			CreatePullRequest(gomock.Any(), gomock.Any()).
			Return(model.PullRequest{}, errors.New("create error"))

		_, err := service.CreatePR(context.Background(), "pr", "name", "author")
		require.Error(t, err)
	})
}

func TestMergePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)

	t.Run("Good: already merged", func(t *testing.T) {
		pr := model.PullRequest{ID: "pr", Status: model.PRStatusMerged}
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)

		result, err := service.MergePR(context.Background(), "pr")
		require.NoError(t, err)
		require.Equal(t, pr, result)
	})

	t.Run("Good: merge now", func(t *testing.T) {
		pr := model.PullRequest{ID: "pr", Status: model.PRStatusOpen}
		merged := pr
		merged.Status = model.PRStatusMerged
		now := time.Now().UTC()

		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			UpdatePullRequestStatus(gomock.Any(), "pr", model.PRStatusMerged, gomock.AssignableToTypeOf(&now)).
			DoAndReturn(func(_ context.Context, _ string, status model.PRStatus, mergedAt *time.Time) (model.PullRequest, error) {
				require.Equal(t, model.PRStatusMerged, status)
				require.NotNil(t, mergedAt)
				return merged, nil
			})

		result, err := service.MergePR(context.Background(), "pr")
		require.NoError(t, err)
		require.Equal(t, merged, result)
	})

	t.Run("Bad: get PR error", func(t *testing.T) {
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "missing").
			Return(model.PullRequest{}, errors.New("not found"))

		_, err := service.MergePR(context.Background(), "missing")
		require.Error(t, err)
	})

	t.Run("Bad: update error", func(t *testing.T) {
		pr := model.PullRequest{ID: "pr", Status: model.PRStatusOpen}
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			UpdatePullRequestStatus(gomock.Any(), "pr", model.PRStatusMerged, gomock.Any()).
			Return(model.PullRequest{}, errors.New("update error"))

		_, err := service.MergePR(context.Background(), "pr")
		require.Error(t, err)
	})
}

func TestReassignReviewer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repository.NewMockRepository(ctrl)
	service := New(repo)

	t.Run("Good: reassign", func(t *testing.T) {
		pr := model.PullRequest{
			ID:        "pr",
			AuthorID:  "author",
			Status:    model.PRStatusOpen,
			Reviewers: []string{"old", "other"},
		}
		reviewer := model.User{ID: "old", TeamID: "team"}
		members := []model.User{
			{ID: "old", TeamID: "team", IsActive: true},
			{ID: "author", TeamID: "team", IsActive: true},
			{ID: "new", TeamID: "team", IsActive: true},
		}
		updated := pr
		updated.Reviewers = []string{"new", "other"}

		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			GetUserByID(gomock.Any(), "old").
			Return(reviewer, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(members, nil)
		repo.EXPECT().
			ReplaceReviewer(gomock.Any(), "pr", "old", "new").
			Return(updated, nil)

		result, replaced, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.NoError(t, err)
		require.Equal(t, "new", replaced)
		require.Equal(t, updated, result)
	})

	t.Run("Bad: get PR error", func(t *testing.T) {
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(model.PullRequest{}, errors.New("get error"))

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.Error(t, err)
	})

	t.Run("Bad: already merged", func(t *testing.T) {
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(model.PullRequest{Status: model.PRStatusMerged}, nil)

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.ErrorIs(t, err, ErrPRMerged)
	})

	t.Run("Bad: reviewer not assigned", func(t *testing.T) {
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(model.PullRequest{Status: model.PRStatusOpen, Reviewers: []string{"other"}}, nil)

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.ErrorIs(t, err, ErrReviewerNotAssigned)
	})

	t.Run("Bad: get reviewer error", func(t *testing.T) {
		pr := model.PullRequest{Status: model.PRStatusOpen, Reviewers: []string{"old"}}
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			GetUserByID(gomock.Any(), "old").
			Return(model.User{}, errors.New("user error"))

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.Error(t, err)
	})

	t.Run("Bad: list members error", func(t *testing.T) {
		pr := model.PullRequest{Status: model.PRStatusOpen, Reviewers: []string{"old"}}
		reviewer := model.User{ID: "old", TeamID: "team"}
		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			GetUserByID(gomock.Any(), "old").
			Return(reviewer, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(nil, errors.New("list error"))

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.Error(t, err)
	})

	t.Run("Bad: no candidates", func(t *testing.T) {
		pr := model.PullRequest{AuthorID: "author", Status: model.PRStatusOpen, Reviewers: []string{"old"}}
		reviewer := model.User{ID: "old", TeamID: "team"}
		members := []model.User{
			{ID: "old", TeamID: "team", IsActive: true},
			{ID: "author", TeamID: "team", IsActive: true},
		}

		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			GetUserByID(gomock.Any(), "old").
			Return(reviewer, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(members, nil)

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.ErrorIs(t, err, ErrNoReplacementCandidate)
	})

	t.Run("Bad: replace error", func(t *testing.T) {
		pr := model.PullRequest{AuthorID: "author", Status: model.PRStatusOpen, Reviewers: []string{"old"}}
		reviewer := model.User{ID: "old", TeamID: "team"}
		members := []model.User{
			{ID: "old", TeamID: "team", IsActive: true},
			{ID: "new", TeamID: "team", IsActive: true},
		}

		repo.EXPECT().
			GetPullRequest(gomock.Any(), "pr").
			Return(pr, nil)
		repo.EXPECT().
			GetUserByID(gomock.Any(), "old").
			Return(reviewer, nil)
		repo.EXPECT().
			ListTeamMembers(gomock.Any(), "team").
			Return(members, nil)
		repo.EXPECT().
			ReplaceReviewer(gomock.Any(), "pr", "old", "new").
			Return(model.PullRequest{}, errors.New("replace error"))

		_, _, err := service.ReassignReviewer(context.Background(), "pr", "old")
		require.Error(t, err)
	})
}
