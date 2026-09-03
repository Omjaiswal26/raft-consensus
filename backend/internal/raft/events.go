package raft

import "time"

const (
	EventElectionStart = "election_start"
	EventRequestVote   = "request_vote"
	EventVoteGranted   = "vote_granted"
	EventBecameLeader  = "became_leader"
	EventHeartbeat     = "heartbeat"
	EventAppendEntries = "append_entries"
)

type Event struct {
	Type string    `json:"type"`
	From uint      `json:"from"`
	To   uint      `json:"to"`
	Term int       `json:"term"`
	At   time.Time `json:"at"`
}

type Emitter interface {
	Emit(e Event)
}
