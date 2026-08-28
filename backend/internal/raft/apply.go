package raft

import (
	"log"
	"strings"
)

func (n *Node) applyCommitted() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.applyCommittedLocked()
}

func (n *Node) applyCommittedLocked() {
	for n.RaftNode.LastApplied < n.RaftNode.CommitIndex {
		n.RaftNode.LastApplied++
		idx := n.RaftNode.LastApplied - 1
		if idx < 0 || idx >= len(n.RaftNode.LogEntries) {
			continue
		}

		entry := n.RaftNode.LogEntries[idx]
		n.applyCommand(entry.Command)
		log.Printf("Node %d applied %q (lastApplied=%d)", n.RaftNode.ID, entry.Command, n.RaftNode.LastApplied)
	}
}

func (n *Node) applyCommand(command string) {
	// v1: only "SET key=value"
	if strings.HasPrefix(command, "SET") {
		rest := strings.TrimPrefix(command, "SET ")
		parts := strings.SplitN(rest, "=", 2)
		if len(parts) == 2 {
			n.KV[parts[0]] = parts[1]
		}
	}
}