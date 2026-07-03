package services

import (
	"raft-consensus/internal/store"
	"raft-consensus/models"
	"time"
)

func InitNodeService(state string, timeout time.Duration) (models.RaftNodeResponse, error) {
	node := models.RaftNode{
		State:       state,
		CurrentTerm: 1,
		VotedFor:    nil,
		CommitIndex: 0,
		LastApplied: 0,
		NextIndex:   0,
		MatchIndex:  0,
		Timeout:     timeout,
		Peers:       []*models.RaftNode{},
	}

	if err := store.CreateNode(&node); err != nil {
		return models.RaftNodeResponse{}, err
	}

	if err := store.UpdateNodePeers(&node); err != nil {
		return models.RaftNodeResponse{}, err
	}

	nodeResponse := models.RaftNodeResponse{
		ID:          node.ID,
		State:       node.State,
		CurrentTerm: node.CurrentTerm,
		VotedFor:    node.VotedFor,
		Timeout:     node.Timeout,
		LeaderID:    node.LeaderID,
		CommitIndex: node.CommitIndex,
		LastApplied: node.LastApplied,
		MatchIndex:  node.MatchIndex,
		Peers:       extractPeerIDs(node.Peers),
	}

	return nodeResponse, nil
}

func extractPeerIDs(peers []*models.RaftNode) []uint {
	var peerIDs []uint

	for _, peer := range peers {
		peerIDs = append(peerIDs, peer.ID)
	}

	return peerIDs
}
