package model

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	Name string
}

type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PRStatus
	Reviewers []string
	CreatedAt time.Time
	MergedAt  *time.Time
}
