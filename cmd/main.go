package main

import (
	"raft-consensus/internal/raft"
	"raft-consensus/models"
)

func main() {
	nodes := []*raft.Node{
		raft.NewNode(&models.RaftNode{State: "follower"}),
		raft.NewNode(&models.RaftNode{State: "follower"}),
		raft.NewNode(&models.RaftNode{State: "follower"}),
	}

	for i, node := range nodes {
		node.Record.ID = uint(i + 1)
	}

	for i, node := range nodes {
		for j := range nodes {
			if i != j {
				node.Record.PeerIDs = append(node.Record.PeerIDs, uint(j+1))
			}
		}
	}

	raft.WireRuntimePeers(nodes)

	for _, node := range nodes {
		node.Start()
	}

	select {}
}
