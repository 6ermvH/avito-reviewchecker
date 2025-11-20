package httpserver

import (
	"net/http"

	"github.com/6ermvH/avito-reviewchecker/internal/model"
	httpmodel "github.com/6ermvH/avito-reviewchecker/internal/model/http"
)

func HandleReviewerStats(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := svc.ListReviewerStats(r.Context())
		if err != nil {
			writeDomainError(w, err, map[string]int{})

			return
		}

		writeJSON(w, http.StatusOK, httpmodel.ReviewerStatsResponse{
			Reviewers: mapReviewerStats(stats),
		})
	}
}

func HandlePullRequestStats(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := svc.GetPullRequestStats(r.Context())
		if err != nil {
			writeDomainError(w, err, map[string]int{})

			return
		}

		writeJSON(w, http.StatusOK, mapPRStats(stats))
	}
}

func mapReviewerStats(stats []model.ReviewerStat) []httpmodel.ReviewerStat {
	resp := make([]httpmodel.ReviewerStat, 0, len(stats))
	for _, stat := range stats {
		resp = append(resp, httpmodel.ReviewerStat{
			UserID:        stat.UserID,
			Username:      stat.Username,
			TeamName:      stat.TeamName,
			TotalAssigned: stat.TotalAssigned,
			OpenAssigned:  stat.OpenAssigned,
		})
	}

	return resp
}

func mapPRStats(stats model.PullRequestStats) httpmodel.PullRequestStatsResponse {
	resp := httpmodel.PullRequestStatsResponse{
		Total:         stats.Total,
		Open:          stats.Open,
		Merged:        stats.Merged,
		AverageReview: stats.AverageReview,
		ByAuthor:      make([]httpmodel.AuthorStat, 0, len(stats.ByAuthor)),
	}

	for _, author := range stats.ByAuthor {
		resp.ByAuthor = append(resp.ByAuthor, httpmodel.AuthorStat{
			AuthorID: author.AuthorID,
			Count:    author.Count,
		})
	}

	return resp
}
