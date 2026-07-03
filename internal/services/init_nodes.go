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
		PeerIDs:     []uint{},
	}

	if err := store.CreateNode(&node); err != nil {
		return models.RaftNodeResponse{}, err
	}

	if err := store.RefreshAllPeerIDs(); err != nil {
		return models.RaftNodeResponse{}, err
	}

	updated, err := store.GetNodeByID(node.ID)
	if err != nil {
		return models.RaftNodeResponse{}, err
	}

	nodeResponse := models.RaftNodeResponse{
		ID:          updated.ID,
		State:       updated.State,
		CurrentTerm: updated.CurrentTerm,
		VotedFor:    updated.VotedFor,
		Timeout:     updated.Timeout,
		LeaderID:    updated.LeaderID,
		CommitIndex: updated.CommitIndex,
		LastApplied: updated.LastApplied,
		MatchIndex:  updated.MatchIndex,
		Peers:       updated.PeerIDs,
	}

	return nodeResponse, nil
}
