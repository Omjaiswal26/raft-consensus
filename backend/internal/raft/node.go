package raft

import (
	"fmt"
	"log"
	"raft-consensus/models"
	"sync"
	"time"
)

type Node struct {
	RaftNode          *models.RaftNode
	Peers             []*Node
	ElectionTimer     *time.Timer
	heartbeatTicker   *time.Ticker
	electionTimeoutCh chan struct{}
	mu                sync.Mutex
	KV                map[string]string `json:"-"`
	emitter           Emitter
}

func NewNode(raftNode *models.RaftNode) *Node {
	return &Node{
		RaftNode:          raftNode,
		electionTimeoutCh: make(chan struct{}, 1),
		KV:                make(map[string]string),
	}
}

func (n *Node) SetEmitter(e Emitter) {
	n.emitter = e
}

func (n *Node) emit(typ string, to uint) {
	if n.emitter == nil {
		return
	}

	n.mu.Lock()
	e := Event{
		Type: typ,
		From: n.RaftNode.ID,
		To:   to,
		Term: n.RaftNode.CurrentTerm,
		At:   time.Now(),
	}
	n.mu.Unlock()
	n.emitter.Emit(e)
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

	if term > n.RaftNode.CurrentTerm {
		n.RaftNode.CurrentTerm = term
		n.RaftNode.VotedFor = nil
	}

	n.RaftNode.State = "follower"
	n.RaftNode.LeaderID = nil

	n.mu.Unlock()

	n.stopHeartbeats()
	n.ResetElectionTimer()
	log.Printf("Node %d became follower for term %d", n.RaftNode.ID, n.RaftNode.CurrentTerm)
}

func (n *Node) BecomeCandidate() {
	n.mu.Lock()

	n.RaftNode.State = "candidate"
	n.RaftNode.CurrentTerm++
	id := int(n.RaftNode.ID)
	n.RaftNode.VotedFor = &id
	n.RaftNode.LeaderID = nil

	term := n.RaftNode.CurrentTerm
	n.mu.Unlock()

	n.stopHeartbeats()
	n.ResetElectionTimer()
	log.Printf("Node %d became candidate for term %d", n.RaftNode.ID, term)

	n.emit(EventElectionStart, 0)

	n.startElection()
}

func (n *Node) startElection() {
	n.mu.Lock()
	term := n.RaftNode.CurrentTerm
	candidateID := int(n.RaftNode.ID)
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

		n.emit(EventRequestVote, peer.RaftNode.ID)

		peer.HandleRequestVote(args, &reply)

		n.mu.Lock()
		if reply.Term > n.RaftNode.CurrentTerm {
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
	stillCandidate := n.RaftNode.State == "candidate" && n.RaftNode.CurrentTerm == term
	n.mu.Unlock()

	if stillCandidate && votes >= majority {
		n.BecomeLeader()
	}
}

func (n *Node) BecomeLeader() {
	n.mu.Lock()
	n.RaftNode.State = "leader"
	leaderID := int(n.RaftNode.ID)
	n.RaftNode.LeaderID = &leaderID
	term := n.RaftNode.CurrentTerm
	n.mu.Unlock()

	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}

	log.Printf("Node %d became leader for term %d", n.RaftNode.ID, term)

	n.emit(EventBecameLeader, 0)

	n.startHeartbeats()
}

func (n *Node) SubmitCommand(command string) error {
	n.mu.Lock()

	if n.RaftNode.State != "leader" {
		n.mu.Unlock()
		return fmt.Errorf("not leader")
	}

	prevLogIndex := n.lastLogIndex()
	prevLogTerm := n.lastLogTerm()

	entry := models.LogEntry{
		Term:    n.RaftNode.CurrentTerm,
		Command: command,
	}

	n.RaftNode.LogEntries = append(n.RaftNode.LogEntries, entry)

	args := AppendEntriesArgs{
		Term:         n.RaftNode.CurrentTerm,
		LeaderID:     int(n.RaftNode.ID),
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      []models.LogEntry{entry},
		LeaderCommit: n.RaftNode.CommitIndex,
	}

	peers := n.Peers
	newIndex := n.lastLogIndex()
	n.mu.Unlock()

	log.Printf("Node %d appended command %q at index %d", n.RaftNode.ID, command, newIndex)

	replicated := 1

	for _, peer := range peers {
		var reply AppendEntriesReply

		n.emit(EventAppendEntries, peer.RaftNode.ID)

		peer.HandleAppendEntries(args, &reply)

		n.mu.Lock()
		if reply.Term > n.RaftNode.CurrentTerm {
			n.mu.Unlock()
			n.stopHeartbeats()
			n.BecomeFollower(reply.Term)
			return fmt.Errorf("stepped down: higher term")
		}
		if reply.Success {
			replicated++
		}
		n.mu.Unlock()
	}

	majority := len(peers)/2 + 1
	if replicated >= majority {
		n.mu.Lock()
		if newIndex > n.RaftNode.CommitIndex {
			n.RaftNode.CommitIndex = newIndex
		}
		n.applyCommittedLocked()
		n.mu.Unlock()
		log.Printf("Node %d committed index %d", n.RaftNode.ID, newIndex)

	}

	return nil
}

func (n *Node) startHeartbeats() {
	n.stopHeartbeats()
	n.heartbeatTicker = time.NewTicker(heartbeatInterval)

	go func() {
		n.broadcastHeartbeat()
		for range n.heartbeatTicker.C {
			n.mu.Lock()
			isLeader := n.RaftNode.State == "leader"
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
	if n.RaftNode.State != "leader" {
		n.mu.Unlock()
		return
	}

	args := AppendEntriesArgs{
		Term:         n.RaftNode.CurrentTerm,
		LeaderID:     int(n.RaftNode.ID),
		PrevLogIndex: n.lastLogIndex(),
		PrevLogTerm:  n.lastLogTerm(),
		Entries:      nil,
		LeaderCommit: n.RaftNode.CommitIndex,
	}
	peers := n.Peers
	n.mu.Unlock()

	for _, peer := range peers {
		var reply AppendEntriesReply
		n.emit(EventHeartbeat, peer.RaftNode.ID)
		peer.HandleAppendEntries(args, &reply)

		n.mu.Lock()
		if reply.Term > n.RaftNode.CurrentTerm {
			n.mu.Unlock()
			n.stopHeartbeats()
			n.BecomeFollower(reply.Term)
			return
		}
		n.mu.Unlock()
	}
}

func (n *Node) lastLogIndex() int {
	return len(n.RaftNode.LogEntries)
}

func (n *Node) lastLogTerm() int {
	if len(n.RaftNode.LogEntries) == 0 {
		return 0
	}

	return n.RaftNode.LogEntries[len(n.RaftNode.LogEntries)-1].Term
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

	reply.Term = n.RaftNode.CurrentTerm
	reply.VoteGranted = false

	// Rule 1: stale term -> reject
	if args.Term < n.RaftNode.CurrentTerm {
		n.mu.Unlock()
		return
	}

	// Rule 2: higher term -> step down to follower
	if args.Term > n.RaftNode.CurrentTerm {
		n.RaftNode.CurrentTerm = args.Term
		n.RaftNode.VotedFor = nil
		n.RaftNode.State = "follower"
		reply.Term = n.RaftNode.CurrentTerm
	}

	// Rule 3: grant if not voted yet (or already voted for them) AND log is fresh enough

	candidateID := args.CandidateID
	if (n.RaftNode.VotedFor == nil || *n.RaftNode.VotedFor == candidateID) && n.isLogUpToDate(args.LastLogIndex, args.LastLogTerm) {
		n.RaftNode.VotedFor = &candidateID
		reply.VoteGranted = true
	}

	granted := reply.VoteGranted
	n.mu.Unlock()

	// Rule 4: granting vote resets election timer (heard from  a valid candidate)

	if granted {
		n.emit(EventVoteGranted, uint(candidateID))
		n.ResetElectionTimer()
	}
}

func (n *Node) HandleAppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	n.mu.Lock()

	reply.Term = n.RaftNode.CurrentTerm
	reply.Success = false

	if args.Term < n.RaftNode.CurrentTerm {
		n.mu.Unlock()
		return
	}

	if args.Term > n.RaftNode.CurrentTerm {
		n.RaftNode.CurrentTerm = args.Term
		n.RaftNode.VotedFor = nil
	}

	n.RaftNode.State = "follower"
	leaderID := args.LeaderID
	n.RaftNode.LeaderID = &leaderID
	reply.Term = n.RaftNode.CurrentTerm

	if args.PrevLogIndex > 0 {
		if args.PrevLogIndex > len(n.RaftNode.LogEntries) {
			n.mu.Unlock()
			return
		}
		if n.RaftNode.LogEntries[args.PrevLogIndex-1].Term != args.PrevLogTerm {
			n.mu.Unlock()
			return
		}
	} else if args.PrevLogTerm != 0 {
		n.mu.Unlock()
		return
	}

	if len(args.Entries) > 0 {
		n.RaftNode.LogEntries = append(n.RaftNode.LogEntries[:args.PrevLogIndex], args.Entries...)
	}

	if args.LeaderCommit > n.RaftNode.CommitIndex {
		lastNewIndex := len(n.RaftNode.LogEntries)
		if args.LeaderCommit < lastNewIndex {
			n.RaftNode.CommitIndex = args.LeaderCommit
		} else {
			n.RaftNode.CommitIndex = lastNewIndex
		}
	}

	n.applyCommittedLocked()

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
	state := n.RaftNode.State
	n.mu.Unlock()

	switch state {
	case "follower", "candidate":
		n.BecomeCandidate()
		// leader: do nothing
	}
}
