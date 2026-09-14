package raft

import (
	"raft-consensus/models"
	"testing"
)

func TestIsLogUpToDate_termThenIndex(t *testing.T) {
	n := NewNode(&models.RaftNode{
		ID: 1,
		LogEntries: []models.LogEntry{
			{Term: 1, Command: "SET a=1"},
			{Term: 2, Command: "SET b=2"},
		},
	})

	cases := []struct {
		name      string
		idx, term int
		want      bool
	}{
		{"newer term wins", 1, 3, true},
		{"older term loses", 5, 1, false},
		{"same term shorter loses", 1, 2, false},
		{"same term equal ok", 2, 2, true},
		{"same term longer ok", 3, 2, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := n.isLogUpToDate(tc.idx, tc.term); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
