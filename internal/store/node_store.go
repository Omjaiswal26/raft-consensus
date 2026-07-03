package store

import (
	"fmt"
	"raft-consensus/database"
	"raft-consensus/models"
)

func CreateNode(node *models.RaftNode) error {
	if err := database.DB.Create(node).Error; err != nil {
		return fmt.Errorf("failed to save node in database: %w", err)
	}
	return nil
}

func UpdateNodePeers(node *models.RaftNode) error {
	var allNodes []models.RaftNode

	if err := database.DB.Find(&allNodes).Error; err != nil {
		return fmt.Errorf("failed to find all nodes: %w", err)
	}

	var peersList []*models.RaftNode
	for i := range allNodes {
		if allNodes[i].ID != node.ID {
			peersList = append(peersList, &allNodes[i])
		}
	}

	node.Peers = peersList

	if err := database.DB.Save(node).Error; err != nil {
		return fmt.Errorf("failed to update node's peers: %w", err)
	}

	return nil
}
