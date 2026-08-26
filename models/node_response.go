package models

import "time"

type RaftNodeResponse struct {
	ID          uint          `json:"id"`
	State       string        `json:"state"`
	CurrentTerm int           `json:"current_term"`
	VotedFor    *int          `json:"voted_for"`
	Timeout     time.Duration `json:"timeout"`
	LeaderID    *int          `json:"leader_id"`
	CommitIndex int           `json:"commit_index"`
	LastApplied int           `json:"last_applied"`
	MatchIndex  int           `json:"match_index"`
	Peers       []uint        `json:"peers"`
	Log         []LogEntry    `json:"log"`
}
