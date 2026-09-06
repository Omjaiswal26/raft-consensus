package services

import (
	"fmt"
	"raft-consensus/internal/dto"
	"raft-consensus/internal/raft"
)

type ClusterService struct {
	nodes []*raft.Node
}

func NewClusterService(nodes []*raft.Node) *ClusterService {
	return &ClusterService{nodes: nodes}
}

func (s *ClusterService) Snapshot() []dto.RaftNodeResponse {
	out := make([]dto.RaftNodeResponse, 0, len(s.nodes))

	for _, n := range s.nodes {
		out = append(out, n.Snapshot())
	}

	return out
}

func (s *ClusterService) SubmitCommand(command string) error {
	for _, n := range s.nodes {
		snap := n.Snapshot()
		if snap.State == "leader" {
			return n.SubmitCommand(command)
		}
	}
	return fmt.Errorf("no leader")
}

func (s *ClusterService) CrashNode(nodeID uint) error {
	nodes := s.nodes

	var reqNode *raft.Node

	for _, node := range nodes {
		if node.RaftNode.ID == nodeID {
			reqNode = node
			break
		}
	}

	if reqNode == nil {
		return fmt.Errorf("node %d not found", nodeID)
	}

	reqNode.Alive = false
	return nil
}

func (s *ClusterService) RecoverNode(nodeID uint) error {
	nodes := s.nodes

	var reqNode *raft.Node

	for _, node := range nodes {
		if node.RaftNode.ID == nodeID {
			reqNode = node
			break
		}
	}

	if reqNode == nil {
		return fmt.Errorf("node %d not found", nodeID)
	}

	reqNode.Alive = true
	return nil
}