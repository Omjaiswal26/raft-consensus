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
		raft.NewNode(&models.RaftNode{ID: 1, State: "follower"}),
		raft.NewNode(&models.RaftNode{ID: 2, State: "follower"}),
		raft.NewNode(&models.RaftNode{ID: 3, State: "follower"}),
	}


	for i, node := range nodes {
		for j := range nodes {
			if i != j {
				node.RaftNode.PeerIDs = append(node.RaftNode.PeerIDs, uint(j+1))
			}
		}
	}

	raft.WireRuntimePeers(nodes)

	hub := api.NewHub()
	for _, node := range nodes {
		node.SetEmitter(hub)
	}

	for _, node := range nodes {
		node.Start()
	}

	time.Sleep(5 * time.Second)
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
	h := api.NewClusterHandler(svc, hub)
	r := api.NewRouter(h)
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
