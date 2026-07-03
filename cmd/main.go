package main

import (
	"raft-consensus/models"
	"raft-consensus/internal/raft"
)

func main() {
	nodes := []*raft.Node{
		raft.NewNode(&models.RaftNode{State: "follower"}),
		raft.NewNode(&models.RaftNode{State: "follower"}),
		raft.NewNode(&models.RaftNode{State: "follower"}),
	}

	for i, node := range nodes {
		node.Record.ID = uint(i + 1)
		node.Start()
	}

	select {}
}
