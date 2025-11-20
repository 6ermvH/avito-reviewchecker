package model

type ReviewerStat struct {
	UserID        string
	Username      string
	TeamName      string
	TotalAssigned int
	OpenAssigned  int
}

type AuthorStat struct {
	AuthorID string
	Count    int
}

type PullRequestStats struct {
	Total         int
	Open          int
	Merged        int
	AverageReview float64
	ByAuthor      []AuthorStat
}
