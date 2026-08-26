package services

import (
	"fmt"
	"raft-consensus/internal/raft"
	"raft-consensus/models"
)

type ClusterService struct {
	nodes []*raft.Node
}

func NewClusterService(nodes []*raft.Node) *ClusterService {
	return &ClusterService{nodes: nodes}
}

func (s *ClusterService) Snapshot() []models.RaftNodeResponse {
	out := make([]models.RaftNodeResponse, 0, len(s.nodes))

	for _, n := range s.nodes {
		out = append(out, n.Snapshot())
	}
	
	return out
}

func (s *ClusterService) SubmitCommand(command string) error {
	for _, n := range s.nodes {
		snap := n.Snapshot()
		if snap.State == "leader" {
			return  n.SubmitCommand(command)
		}
	}
	return fmt.Errorf("no leader")
}