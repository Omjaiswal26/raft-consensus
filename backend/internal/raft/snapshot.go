package raft

import (
	"maps"
	"raft-consensus/internal/dto"
	"raft-consensus/models"
)

func (n *Node) Snapshot() dto.RaftNodeResponse {
	n.mu.Lock()
	defer n.mu.Unlock()

	kvCopy := make(map[string]string, len(n.KV))
	for k, v := range n.KV {
		kvCopy[k] = v
	}

	var nextCopy, matchCopy map[uint]uint
	if n.RaftNode.State == "leader" {
		nextCopy = maps.Clone(n.nextIndex)
		matchCopy = maps.Clone(n.matchIndex)
	}

	return dto.RaftNodeResponse{
		ID:          n.RaftNode.ID,
		State:       n.RaftNode.State,
		CurrentTerm: n.RaftNode.CurrentTerm,
		VotedFor:    n.RaftNode.VotedFor,
		Timeout:     n.RaftNode.Timeout,
		LeaderID:    n.RaftNode.LeaderID,
		CommitIndex: n.RaftNode.CommitIndex,
		LastApplied: n.RaftNode.LastApplied,
		Peers:       append([]uint(nil), n.RaftNode.PeerIDs...),
		Log:         append([]models.LogEntry(nil), n.RaftNode.LogEntries...),
		KV:          kvCopy,
		NextIndex:   nextCopy,
		MatchIndex:  matchCopy,
	}
}
