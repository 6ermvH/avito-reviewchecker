package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
)

func (r *Repository) ListReviewerStats(ctx context.Context) ([]model.ReviewerStat, error) {
	query := `
SELECT u.id,
       u.username,
       u.team_name,
       COUNT(*) AS total_assigned,
       COALESCE(SUM(CASE WHEN pr.status = 'OPEN' THEN 1 ELSE 0 END), 0) AS open_assigned
FROM pull_request_reviewers r
JOIN users u ON u.id = r.reviewer_id
JOIN pull_requests pr ON pr.id = r.pull_request_id
GROUP BY u.id, u.username, u.team_name
ORDER BY total_assigned DESC, u.id
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list reviewer stats: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	var stats []model.ReviewerStat

	for rows.Next() {
		var stat model.ReviewerStat
		if err := rows.Scan(&stat.UserID, &stat.Username,
			&stat.TeamName, &stat.TotalAssigned,
			&stat.OpenAssigned); err != nil {
			return nil, fmt.Errorf("scan reviewer stat: %w", err)
		}

		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("list reviewer stats: %w", err)
	}

	return stats, nil
}

func (r *Repository) GetPullRequestStats(ctx context.Context) (model.PullRequestStats, error) {
	type aggregate struct {
		Total  sql.NullInt64
		Open   sql.NullInt64
		Merged sql.NullInt64
	}

	row := r.db.QueryRowContext(ctx, `
SELECT COUNT(*) AS total,
       SUM(CASE WHEN status = 'OPEN' THEN 1 ELSE 0 END) AS open,
       SUM(CASE WHEN status = 'MERGED' THEN 1 ELSE 0 END) AS merged
FROM pull_requests
`)

	var agg aggregate
	if err := row.Scan(&agg.Total, &agg.Open, &agg.Merged); err != nil {
		return model.PullRequestStats{}, fmt.Errorf("pull request aggregates: %w", err)
	}

	totalPR := int(agg.Total.Int64)

	var totalAssignments int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pull_request_reviewers`).
		Scan(&totalAssignments); err != nil {
		return model.PullRequestStats{}, fmt.Errorf("count reviewer assignments: %w", err)
	}

	stats := model.PullRequestStats{
		Total:  totalPR,
		Open:   int(agg.Open.Int64),
		Merged: int(agg.Merged.Int64),
	}
	if totalPR > 0 {
		stats.AverageReview = float64(totalAssignments) / float64(totalPR)
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT author_id, COUNT(*) AS count
FROM pull_requests
GROUP BY author_id
ORDER BY count DESC, author_id
`)
	if err != nil {
		return stats, fmt.Errorf("pull request stats by author: %w", err)
	}
	//nolint:errcheck
	defer rows.Close()

	for rows.Next() {
		var stat model.AuthorStat
		if err := rows.Scan(&stat.AuthorID, &stat.Count); err != nil {
			return stats, fmt.Errorf("scan author stat: %w", err)
		}

		stats.ByAuthor = append(stats.ByAuthor, stat)
	}

	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("iterate author stats: %w", err)
	}

	return stats, nil
}
