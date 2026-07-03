package raft

import (
	"log"
	"raft-consensus/models"
	"sync"
	"time"
)

type Node struct {
	Record            *models.RaftNode
	Peers             []*Node
	ElectionTimer     *time.Timer
	heartbeatTicker   *time.Ticker
	electionTimeoutCh chan struct{}
	mu                sync.Mutex
}

func NewNode(record *models.RaftNode) *Node {
	return &Node{Record: record, electionTimeoutCh: make(chan struct{}, 1)}
}

func (n *Node) ResetElectionTimer() {
	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}

	d := RandomElectionTimeout()
	n.ElectionTimer = time.AfterFunc(d, func() {
		select {
		case n.electionTimeoutCh <- struct{}{}:
		default:
		}
	})
}

func (n *Node) stopHeartbeats() {
	if n.heartbeatTicker != nil {
		n.heartbeatTicker.Stop()
		n.heartbeatTicker = nil
	}
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

	n.stopHeartbeats()
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

	n.stopHeartbeats()
	n.ResetElectionTimer()
	log.Printf("Node %d became candidate for term %d", n.Record.ID, term)

	// Request votes from peers
	n.startElection()
}

func (n *Node) startElection() {
	n.mu.Lock()
	term := n.Record.CurrentTerm
	candidateID := int(n.Record.ID)
	lastIndex := n.lastLogIndex()
	lastTerm := n.lastLogTerm()
	peers := n.Peers
	n.mu.Unlock()

	votes := 1

	args := RequestVoteArgs{
		Term:         term,
		CandidateID:  candidateID,
		LastLogIndex: lastIndex,
		LastLogTerm:  lastTerm,
	}

	for _, peer := range peers {
		var reply RequestVoteReply
		peer.HandleRequestVote(args, &reply)

		n.mu.Lock()
		if reply.Term > n.Record.CurrentTerm {
			n.mu.Unlock()
			n.BecomeFollower(reply.Term)
			return
		}

		if reply.VoteGranted {
			votes++
		}

		n.mu.Unlock()
	}

	majority := len(peers)/2 + 1

	n.mu.Lock()
	stillCandidate := n.Record.State == "candidate" && n.Record.CurrentTerm == term
	n.mu.Unlock()

	if stillCandidate && votes >= majority {
		n.BecomeLeader()
	}
}

func (n *Node) BecomeLeader() {
	n.mu.Lock()
	n.Record.State = "leader"
	leaderID := int(n.Record.ID)
	n.Record.LeaderID = &leaderID
	term := n.Record.CurrentTerm
	n.mu.Unlock()

	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}

	log.Printf("Node %d became leader for term %d", n.Record.ID, term)
	n.startHeartbeats()
}

func (n *Node) startHeartbeats() {
	n.stopHeartbeats()
	n.heartbeatTicker = time.NewTicker(heartbeatInterval)

	go func() {
		n.broadcastHeartbeat()
		for range n.heartbeatTicker.C {
			n.mu.Lock()
			isLeader := n.Record.State == "leader"
			n.mu.Unlock()
			if !isLeader {
				return
			}
			n.broadcastHeartbeat()
		}
	}()
}

func (n *Node) broadcastHeartbeat() {
	n.mu.Lock()
	if n.Record.State != "leader" {
		n.mu.Unlock()
		return
	}

	args := AppendEntriesArgs{
		Term:         n.Record.CurrentTerm,
		LeaderID:     int(n.Record.ID),
		PrevLogIndex: n.lastLogIndex(),
		PrevLogTerm:  n.lastLogTerm(),
		Entries:      nil,
		LeaderCommit: n.Record.CommitIndex,
	}
	peers := n.Peers
	n.mu.Unlock()

	for _, peer := range peers {
		var reply AppendEntriesReply
		peer.HandleAppendEntries(args, &reply)

		n.mu.Lock()
		if reply.Term > n.Record.CurrentTerm {
			n.mu.Unlock()
			n.stopHeartbeats()
			n.BecomeFollower(reply.Term)
			return
		}
		n.mu.Unlock()
	}
}

func (n *Node) lastLogIndex() int {
	return len(n.Record.LogEntries)
}

func (n *Node) lastLogTerm() int {
	if len(n.Record.LogEntries) == 0 {
		return 0
	}

	return n.Record.LogEntries[len(n.Record.LogEntries)-1].Term
}

func (n *Node) isLogUpToDate(candidateLastIndex, candidateLastTerm int) bool {
	ourTerm := n.lastLogTerm()
	if candidateLastTerm != ourTerm {
		return candidateLastTerm > ourTerm
	}

	return candidateLastIndex >= n.lastLogIndex()
}

func (n *Node) HandleRequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	n.mu.Lock()

	reply.Term = n.Record.CurrentTerm
	reply.VoteGranted = false

	// Rule 1: stale term -> reject
	if args.Term < n.Record.CurrentTerm {
		n.mu.Unlock()
		return
	}

	// Rule 2: higher term -> step down to follower
	if args.Term > n.Record.CurrentTerm {
		n.Record.CurrentTerm = args.Term
		n.Record.VotedFor = nil
		n.Record.State = "follower"
		reply.Term = n.Record.CurrentTerm
	}

	// Rule 3: grant if not voted yet (or already voted for them) AND log is fresh enough

	candidateID := args.CandidateID
	if (n.Record.VotedFor == nil || *n.Record.VotedFor == candidateID) && n.isLogUpToDate(args.LastLogIndex, args.LastLogTerm) {
		n.Record.VotedFor = &candidateID
		reply.VoteGranted = true
	}

	granted := reply.VoteGranted
	n.mu.Unlock()

	// Rule 4: granting vote resets election timer (heard from  a valid candidate)

	if granted {
		n.ResetElectionTimer()
	}
}

func (n *Node) HandleAppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	n.mu.Lock()

	reply.Term = n.Record.CurrentTerm
	reply.Success = false

	if args.Term < n.Record.CurrentTerm {
		n.mu.Unlock()
		return
	}

	if args.Term > n.Record.CurrentTerm {
		n.Record.CurrentTerm = args.Term
		n.Record.VotedFor = nil
	}

	n.Record.State = "follower"
	leaderID := args.LeaderID
	n.Record.LeaderID = &leaderID
	reply.Term = n.Record.CurrentTerm

	if args.PrevLogIndex > 0 {
		if args.PrevLogIndex > len(n.Record.LogEntries) {
			n.mu.Unlock()
			return
		}
		if n.Record.LogEntries[args.PrevLogIndex-1].Term != args.PrevLogTerm {
			n.mu.Unlock()
			return
		}
	} else if args.PrevLogTerm != 0 {
		n.mu.Unlock()
		return
	}

	if len(args.Entries) > 0 {
		n.Record.LogEntries = append(n.Record.LogEntries[:args.PrevLogIndex], args.Entries...)
	}

	if args.LeaderCommit > n.Record.CommitIndex {
		lastNewIndex := len(n.Record.LogEntries)
		if args.LeaderCommit < lastNewIndex {
			n.Record.CommitIndex = args.LeaderCommit
		} else {
			n.Record.CommitIndex = lastNewIndex
		}
	}

	reply.Success = true
	n.mu.Unlock()

	n.ResetElectionTimer()
}

func (n *Node) Start() {
	go func() {
		n.BecomeFollower(0)
		for range n.electionTimeoutCh {
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
		// leader: do nothing
	}
}
