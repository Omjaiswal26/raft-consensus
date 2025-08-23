package services

import (
	"fmt"
	"raft-consensus/database"
	"raft-consensus/models"
	"time"
)


func InitNodeService(state string, timeout time.Duration) (models.RaftNodeResponse, error){
	node := models.RaftNode{
		State:				state,
		CurrentTerm: 		1,
		VotedFor: 			nil,
		CommitIndex: 		0,
		LastApplied:  		0,
		NextIndex: 			0,
		MatchIndex: 		0,
		Timeout: 			timeout,
		HeartbeatTicker: 	nil,
		Peers: 				[]*models.RaftNode{},
	}

	if err := database.DB.Create(&node).Error; err != nil {
		return models.RaftNodeResponse{}, fmt.Errorf("Failed to save node in database: %v", err)
	}

	if err := node.UpdateNodePeers(); err != nil {
		return models.RaftNodeResponse{}, err
	}

	nodeResponse := models.RaftNodeResponse{
		ID:            node.ID,
		State:         node.State,
		CurrentTerm:   node.CurrentTerm,
		VotedFor:      node.VotedFor,
		Timeout:       node.Timeout,
		LeaderID:      node.LeaderID,
		CommitIndex:   node.CommitIndex,
		LastApplied:   node.LastApplied,
		MatchIndex:    node.MatchIndex,
		Peers:         extractPeerIDs(node.Peers), // Optionally, just return peer IDs
	}

	return nodeResponse, nil
}


func extractPeerIDs(peers []*models.RaftNode) ([]uint) {
	var peerIDs []uint

	for _, peer := range peers {
		peerIDs = append(peerIDs, peer.ID)
	}

	return peerIDs
}