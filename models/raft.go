package models

import (
	"fmt"
	"raft-consensus/database"
	"sync"
	"time"

	"gorm.io/gorm"
)

type LogEntry struct {
	Term	int		`json:"term"`
	Command	string	`json:"command"`
}


type RaftNode struct {
	gorm.Model
	State 			string 			`json:"state"`  //Leader, Candidate, Follower
	CurrentTerm		int 			`json:"current_term"`
	VotedFor		*int 			`json:"voted_for"`
	LogEntries		[]LogEntry		`json:"log_entries"`
	CommitIndex		int				`json:"commit_index"`
	LastApplied		int				`json:"last_applied"`
	NextIndex		int				`json:"next_index"`
	MatchIndex		int				`json:"match_index"`
	Peers			[]*RaftNode		`json:"peers"`
	LeaderID		*int			`json:"leader_id"`
	HeartbeatTicker	*time.Ticker	`json:"heartbeat_ticker"`
	Timeout			time.Duration	`json:"timeout"`
	Mutex			sync.Mutex
}

func (node *RaftNode) UpdateNodePeers() (error) {
	var allNodes []RaftNode

	if err := database.DB.Find(&allNodes).Error; err != nil {
		return fmt.Errorf("Failed to find all nodes: %v", err)
	}

	var peersList []*RaftNode
	for i := range allNodes {
		if allNodes[i].ID != node.ID {
			peersList = append(peersList, &allNodes[i])
		}
	}

	node.Peers = peersList

	if err := database.DB.Save(node).Error; err != nil {
		return fmt.Errorf("Failed to update node's peers: %v", err)
	}

	return nil
}