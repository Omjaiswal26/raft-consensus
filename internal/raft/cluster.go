package raft

// WireRuntimePeers links in-memory peer pointers from persisted peer IDs.
func WireRuntimePeers(nodes []*Node) {
	byID := make(map[uint]*Node, len(nodes))
	for _, n := range nodes {
		byID[n.RaftNode.ID] = n
	}

	for _, n := range nodes {
		n.Peers = nil
		for _, peerID := range n.RaftNode.PeerIDs {
			if peer, ok := byID[peerID]; ok {
				n.Peers = append(n.Peers, peer)
			}
		}
	}
}
