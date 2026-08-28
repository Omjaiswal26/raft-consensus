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

func LoadAllNodes() ([]models.RaftNode, error) {
	var nodes []models.RaftNode
	if err := database.DB.Find(&nodes).Error; err != nil {
		return nil, fmt.Errorf("failed to find all nodes: %w", err)
	}
	return nodes, nil
}

func GetNodeByID(id uint) (*models.RaftNode, error) {
	var node models.RaftNode
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find node %d: %w", id, err)
	}
	return &node, nil
}

func UpdateNodePeerIDs(node *models.RaftNode) error {
	allNodes, err := LoadAllNodes()
	if err != nil {
		return err
	}

	var peerIDs []uint
	for _, other := range allNodes {
		if other.ID != node.ID {
			peerIDs = append(peerIDs, other.ID)
		}
	}

	node.PeerIDs = peerIDs
	if err := database.DB.Save(node).Error; err != nil {
		return fmt.Errorf("failed to update node's peer IDs: %w", err)
	}
	return nil
}

func RefreshAllPeerIDs() error {
	allNodes, err := LoadAllNodes()
	if err != nil {
		return err
	}

	for i := range allNodes {
		if err := UpdateNodePeerIDs(&allNodes[i]); err != nil {
			return err
		}
	}
	return nil
}
