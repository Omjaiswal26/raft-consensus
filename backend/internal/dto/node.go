package dto

import (
	"raft-consensus/models"
	"time"
)

type RaftNodeResponse struct {
	ID          uint              `json:"id"`
	State       string            `json:"state"`
	CurrentTerm int               `json:"current_term"`
	VotedFor    *int              `json:"voted_for"`
	Timeout     time.Duration     `json:"timeout"`
	LeaderID    *int              `json:"leader_id"`
	CommitIndex int               `json:"commit_index"`
	LastApplied int               `json:"last_applied"`
	Peers       []uint            `json:"peers"`
	Log         []models.LogEntry `json:"log"`
	KV          map[string]string `json:"kv"`
	NextIndex   map[uint]uint     `json:"next_index,omitempty"`
	MatchIndex  map[uint]uint     `json:"match_index,omitempty"`
}
