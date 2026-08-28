package main

import (
	"log"
	"raft-consensus/internal/api"
	"raft-consensus/internal/raft"
	"raft-consensus/internal/services"
	"raft-consensus/models"
	"time"
)

func main() {
	nodes := []*raft.Node{
		raft.NewNode(&models.RaftNode{State: "follower"}),
		raft.NewNode(&models.RaftNode{State: "follower"}),
		raft.NewNode(&models.RaftNode{State: "follower"}),
	}

	for i, node := range nodes {
		node.RaftNode.ID = uint(i + 1)
	}

	for i, node := range nodes {
		for j := range nodes {
			if i != j {
				node.RaftNode.PeerIDs = append(node.RaftNode.PeerIDs, uint(j+1))
			}
		}
	}

	raft.WireRuntimePeers(nodes)

	for _, node := range nodes {
		node.Start()
	}

	time.Sleep(500 * time.Millisecond)
	var leader *raft.Node
	for _, node := range nodes {
		if node.RaftNode.State == "leader" {
			leader = node
			break
		}
	}
	if leader == nil {
		log.Fatal("no leader")
	}
	if err := leader.SubmitCommand("SET x=1"); err != nil {
		log.Fatal(err)
	}
	for _, node := range nodes {
		log.Printf("node %d log=%v commit=%d",
			node.RaftNode.ID, node.RaftNode.LogEntries, node.RaftNode.CommitIndex)
	}

	svc := services.NewClusterService(nodes)
	h := api.NewClusterHandler(svc)
	r := api.NewRouter(h)
	r.Run()
}
