package services

import (
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