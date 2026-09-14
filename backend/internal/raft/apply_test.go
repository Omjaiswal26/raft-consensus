package raft

import (
	"raft-consensus/models"
	"testing"
)

func TestApplyCommand_SET(t *testing.T) {
	n := NewNode(&models.RaftNode{ID: 1, State: "follower"})

	n.applyCommand("SET x=1")

	got := n.KV["x"]
	want := "1"

	if got != want {
		t.Fatalf("KV[x] = %q, want %q", got, want)
	}
}

func TestApplyCommand_ignoresNonSet(t *testing.T) {
	n := NewNode(&models.RaftNode{ID: 1, State: "follower"})
	n.applyCommand("GET x")

	if len(n.KV) != 0 {
		t.Fatalf("expected empty KV, got %v", n.KV)
	}
}

func TestApplyCommittedLocked(t *testing.T) {
	n := NewNode(&models.RaftNode{
		ID: 1, State: "follower",
		LogEntries:  []models.LogEntry{{Term: 1, Command: "SET x=1"}},
		CommitIndex: 1,
		LastApplied: 0,
	})

	n.applyCommittedLocked()

	if n.RaftNode.LastApplied != 1 {
		t.Fatalf("LastApplied=%d, want 1", n.RaftNode.LastApplied)
	}

	if n.KV["x"] != "1" {
		t.Fatalf("KV[x]=%q, want 1", n.KV["x"])
	}
}

func TestMaybeCommitLocked(t *testing.T) {
	leader := NewNode(&models.RaftNode{
		ID: 1, State: "leader", CurrentTerm: 1,
		LogEntries: []models.LogEntry{{Term: 1, Command: "SET x=1"}},
		CommitIndex: 0,
	})

	f2 := NewNode(&models.RaftNode{ID: 2, State: "follower"})
	f3 := NewNode(&models.RaftNode{ID: 3, State: "follower"})

	leader.Peers = []*Node{f2, f3}
	leader.matchIndex = map[uint]uint{
		2: 1,
		3: 0,
	}

	leader.maybeCommitLocked()

	if leader.RaftNode.CommitIndex != 1 {
		t.Fatalf("CommitIndex=%d, want 1", leader.RaftNode.CommitIndex)
	}

	if leader.KV["x"] != "1" {
		t.Fatalf("KV[x]=%q, want 1", leader.KV["x"])
	}
}

func TestMaybeCommitLocked_noMajority(t *testing.T) {
	leader := NewNode(&models.RaftNode{
		ID: 1, State: "leader", CurrentTerm: 1,
		LogEntries: []models.LogEntry{{Term: 1, Command: "SET x=1"}},
		CommitIndex: 0,
	})

	f2 := NewNode(&models.RaftNode{ID: 2, State: "follower"})
	f3 := NewNode(&models.RaftNode{ID: 3, State: "follower"})

	leader.Peers = []*Node{f2, f3}

	leader.matchIndex = map[uint]uint {
		2: 0,
		3: 0,
	}

	leader.maybeCommitLocked()

	if leader.RaftNode.CommitIndex != 0 {
		t.Fatalf("CommitIndex=%d, want 0", leader.RaftNode.CommitIndex)
	}

	if len(leader.KV) != 0 {
		t.Fatalf("expected empty KV, got %v", leader.KV)
	}
}
