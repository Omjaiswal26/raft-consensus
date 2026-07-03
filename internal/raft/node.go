package raft

import (
	"log"
	"raft-consensus/models"
	"sync"
	"time"
)

type Node struct {
	Record        *models.RaftNode
	ElectionTimer *time.Timer
	mu            sync.Mutex
}

func NewNode(record *models.RaftNode) *Node {
	return &Node{Record: record}
}

func (n *Node) ResetElectionTimer() {
	if n.ElectionTimer != nil {
		if !n.ElectionTimer.Stop() {
			select {
			case <-n.ElectionTimer.C:
			default:
			}
		}
	}

	n.ElectionTimer = time.NewTimer(RandomElectionTimeout())
}

func (n *Node) BecomeFollower(term int) {
	n.mu.Lock()

	if term > n.Record.CurrentTerm {
		n.Record.CurrentTerm = term
		n.Record.VotedFor = nil
	}

	n.Record.State = "follower"
	n.Record.LeaderID = nil

	n.mu.Unlock()

	n.ResetElectionTimer()
	log.Printf("Node %d became follower for term %d", n.Record.ID, n.Record.CurrentTerm)
}

func (n *Node) BecomeCandidate() {
	n.mu.Lock()

	n.Record.State = "candidate"
	n.Record.CurrentTerm++
	id := int(n.Record.ID)
	n.Record.VotedFor = &id
	n.Record.LeaderID = nil

	term := n.Record.CurrentTerm
	n.mu.Unlock()

	n.ResetElectionTimer()
	log.Printf("Node %d became candidate for term %d", n.Record.ID, term)

	// Request votes from peers
	n.startElection()
}

func (n *Node) startElection() {
	// TODO: send RequestVote RPCs to all peers
	log.Printf("Node %d would request votes for term %d", n.Record.ID, n.Record.CurrentTerm)
}


func (n *Node) Start() {
	go func() {
		n.BecomeFollower(0)

		for {
			n.mu.Lock()
			timer := n.ElectionTimer
			n.mu.Unlock()

			<-timer.C
			n.onElectionTimeout()
		}
	}()
}

func (n *Node) onElectionTimeout() {
	n.mu.Lock()
	state := n.Record.State
	n.mu.Unlock()

	switch state {
	case "follower", "candidate":
		n.BecomeCandidate()
	}
}