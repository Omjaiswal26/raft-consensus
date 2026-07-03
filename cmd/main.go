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
		node.Record.ID = uint(i+1)
		for j, other := range nodes {
			if i != j {
				node.Peers = append(node.Peers, other)
			}
		}
	}
	for _, node := range nodes {
		node.Start()
	}

	select {}
}
