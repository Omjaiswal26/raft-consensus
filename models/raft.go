package models

import (
	"time"

	"gorm.io/gorm"
)

type LogEntry struct {
	Term    int    `json:"term"`
	Command string `json:"command"`
}

type RaftNode struct {
	gorm.Model
	State       string        `json:"state"` // Leader, Candidate, Follower
	CurrentTerm int           `json:"current_term"`
	VotedFor    *int          `json:"voted_for"`
	LogEntries  []LogEntry    `json:"log_entries"`
	CommitIndex int           `json:"commit_index"`
	LastApplied int           `json:"last_applied"`
	NextIndex   int           `json:"next_index"`
	MatchIndex  int           `json:"match_index"`
	Peers       []*RaftNode   `json:"peers"`
	LeaderID    *int          `json:"leader_id"`
	Timeout     time.Duration `json:"timeout"`
}
