package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
	"github.com/6ermvH/avito-reviewchecker/internal/repository"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTeamByName(ctx context.Context, name string) (model.Team, error) {
	query := `SELECT name FROM teams WHERE name = $1`

	var team model.Team

	err := r.db.QueryRowContext(ctx, query, name).Scan(&team.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Team{}, repository.ErrNotFound
		}

		return model.Team{}, fmt.Errorf("get team by name, get query row: %w", err)
	}

	return team, nil
}

func (r *Repository) CreateTeam(ctx context.Context, name string) (model.Team, error) {
	query := `INSERT INTO teams (name) VALUES ($1) RETURNING name`

	var team model.Team

	if err := r.db.QueryRowContext(ctx, query, name).Scan(&team.Name); err != nil {
		return model.Team{}, fmt.Errorf("create team, get query row: %w", err)
	}

	return team, nil
}

func (r *Repository) InsertTeamMembers(
	ctx context.Context,
	teamName string,
	users []model.User,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("insert team members, begin transaction: %w", err)
	}

	stmt := `
INSERT INTO users (id, team_name, username, is_active, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (id) DO UPDATE
SET team_name = EXCLUDED.team_name,
    username = EXCLUDED.username,
    is_active = EXCLUDED.is_active,
    updated_at = now()
`
	for _, user := range users {
		if _, err := tx.ExecContext(ctx, stmt, user.ID, teamName, user.Username, user.IsActive); err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("insert team members, exec: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("insert team members, commit: %w", err)
	}

	return nil
}

func (r *Repository) ListTeamMembers(ctx context.Context, teamName string) ([]model.User, error) {
	query := `SELECT id, team_name, username, is_active FROM users WHERE team_name = $1 ORDER BY username`

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("list team members, get query: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var members []model.User

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.TeamName, &user.Username, &user.IsActive); err != nil {
			return nil, fmt.Errorf("list team members, scan user: %w", err)
		}

		members = append(members, user)
	}

	if err := rows.Err(); err != nil {
		return members, fmt.Errorf("list team members, bad rows: %w", err)
	}

	return members, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (model.User, error) {
	query := `SELECT id, team_name, username, is_active FROM users WHERE id = $1`

	var user model.User

	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&user.ID, &user.TeamName, &user.Username, &user.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, repository.ErrNotFound
		}

		return model.User{}, fmt.Errorf("get user, get query row: %w", err)
	}

	return user, nil
}

func (r *Repository) SetUserActivity(
	ctx context.Context,
	userID string,
	active bool,
) (model.User, error) {
	query := `
UPDATE users SET is_active = $1, updated_at = now()
WHERE id = $2
RETURNING id, team_name, username, is_active
`

	var user model.User

	err := r.db.QueryRowContext(ctx, query, active, userID).
		Scan(&user.ID, &user.TeamName, &user.Username, &user.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, repository.ErrNotFound
		}

		return model.User{}, fmt.Errorf("set user active, get query row: %w", err)
	}

	return user, nil
}

func (r *Repository) CreatePullRequest(
	ctx context.Context,
	pr model.PullRequest,
) (model.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("create pr, begin transaction: %w", err)
	}

	insertPR := `
INSERT INTO pull_requests (id, name, author_id, status)
VALUES ($1, $2, $3, $4)
RETURNING id, name, author_id, status, created_at, merged_at
`

	var stored model.PullRequest

	if err := tx.QueryRowContext(ctx, insertPR, pr.ID, pr.Name, pr.AuthorID, pr.Status).
		Scan(&stored.ID, &stored.Name, &stored.AuthorID, &stored.Status, &stored.CreatedAt, &stored.MergedAt); err != nil {
		_ = tx.Rollback()

		return model.PullRequest{}, fmt.Errorf("create pr, get query row: %w", err)
	}

	insertReviewer := `
INSERT INTO pull_request_reviewers (pull_request_id, slot, reviewer_id)
VALUES ($1, $2, $3)
`

	for idx, reviewerID := range pr.Reviewers {
		slot := idx + 1
		if _, err := tx.ExecContext(ctx, insertReviewer, pr.ID, slot, reviewerID); err != nil {
			_ = tx.Rollback()

			return model.PullRequest{}, fmt.Errorf("create pr, exec: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.PullRequest{}, fmt.Errorf("create pr, bad commit: %w", err)
	}

	return r.GetPullRequest(ctx, pr.ID)
}

func (r *Repository) GetPullRequest(ctx context.Context, prID string) (model.PullRequest, error) {
	query := `
SELECT id, name, author_id, status, created_at, merged_at
FROM pull_requests
WHERE id = $1
`

	var pr model.PullRequest

	err := r.db.QueryRowContext(ctx, query, prID).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PullRequest{}, repository.ErrNotFound
		}

		return model.PullRequest{}, fmt.Errorf("get pr, get query row: %w", err)
	}

	reviewers, err := r.loadReviewers(ctx, prID)
	if err != nil {
		return model.PullRequest{}, err
	}

	pr.Reviewers = reviewers

	return pr, nil
}

func (r *Repository) UpdatePullRequestStatus(
	ctx context.Context,
	prID string,
	status model.PRStatus,
	mergedAt *time.Time,
) (model.PullRequest, error) {
	query := `
UPDATE pull_requests
SET status = $1,
    merged_at = $2
WHERE id = $3
RETURNING id, name, author_id, status, created_at, merged_at
`

	var pr model.PullRequest

	if err := r.db.QueryRowContext(ctx, query, status, mergedAt, prID).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PullRequest{}, repository.ErrNotFound
		}

		return model.PullRequest{}, fmt.Errorf("update pr, get query row: %w", err)
	}

	reviewers, err := r.loadReviewers(ctx, prID)
	if err != nil {
		return model.PullRequest{}, err
	}

	pr.Reviewers = reviewers

	return pr, nil
}

func (r *Repository) ListReviewerPullRequests(
	ctx context.Context,
	userID string,
) ([]model.PullRequest, error) {
	query := `
SELECT pr.id, pr.name, pr.author_id, pr.status, pr.created_at, pr.merged_at
FROM pull_requests pr
JOIN pull_request_reviewers r ON pr.id = r.pull_request_id
WHERE r.reviewer_id = $1
`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("load reviewers, get query: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var prs []model.PullRequest

	for rows.Next() {
		var pr model.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, fmt.Errorf("list reviewer, scan pr: %w", err)
		}

		reviewers, err := r.loadReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}

		pr.Reviewers = reviewers
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return prs, fmt.Errorf("list reviewers for user %q: %w", userID, err)
	}

	return prs, nil
}

func (r *Repository) ReplaceReviewer(
	ctx context.Context,
	prID, oldUserID, newUserID string,
) (model.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("bad transaction begin: %w", err)
	}

	query := `
UPDATE pull_request_reviewers
SET reviewer_id = $3,
    assigned_at = now()
WHERE pull_request_id = $1
  AND reviewer_id = $2
`

	res, err := tx.ExecContext(ctx, query, prID, oldUserID, newUserID)
	if err != nil {
		_ = tx.Rollback()

		return model.PullRequest{}, fmt.Errorf("exec in replace reviewer: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return model.PullRequest{}, fmt.Errorf("get affected rows: %w", err)
	}

	if affected == 0 {
		_ = tx.Rollback()

		return model.PullRequest{}, repository.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return model.PullRequest{}, fmt.Errorf("bad commit in replace review: %w", err)
	}

	return r.GetPullRequest(ctx, prID)
}

func (r *Repository) loadReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `
SELECT reviewer_id
FROM pull_request_reviewers
WHERE pull_request_id = $1
ORDER BY slot
`

	rows, err := r.db.QueryContext(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("load reviewers, get query: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var reviewers []string

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("scan reviewerID: %w", err)
		}

		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return reviewers, fmt.Errorf("load reviewers for pr %q: %w", prID, err)
	}

	return reviewers, nil
}
