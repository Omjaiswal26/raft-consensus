package raft

import (
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

	return dto.RaftNodeResponse{
		ID:          n.RaftNode.ID,
		State:       n.RaftNode.State,
		CurrentTerm: n.RaftNode.CurrentTerm,
		VotedFor:    n.RaftNode.VotedFor,
		Timeout:     n.RaftNode.Timeout,
		LeaderID:    n.RaftNode.LeaderID,
		CommitIndex: n.RaftNode.CommitIndex,
		LastApplied: n.RaftNode.LastApplied,
		MatchIndex:  n.RaftNode.MatchIndex,
		Peers:       append([]uint(nil), n.RaftNode.PeerIDs...),
		Log:         append([]models.LogEntry(nil), n.RaftNode.LogEntries...),
		KV:          kvCopy,
	}
}
