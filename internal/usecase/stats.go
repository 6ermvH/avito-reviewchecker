package usecase

import (
	"context"
	"fmt"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
)

func (s *Service) ListReviewerStats(ctx context.Context) ([]model.ReviewerStat, error) {
	s.logger.Debug("list reviewer stats")

	stats, err := s.repo.ListReviewerStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("list reviewer stats: %w", err)
	}

	return stats, nil
}

func (s *Service) GetPullRequestStats(ctx context.Context) (model.PullRequestStats, error) {
	s.logger.Debug("get pull request stats")

	stats, err := s.repo.GetPullRequestStats(ctx)
	if err != nil {
		return model.PullRequestStats{}, fmt.Errorf("get pull request stats: %w", err)
	}

	return stats, nil
}
