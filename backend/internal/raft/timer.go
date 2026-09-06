package raft 

import (
	"math/rand"
	"time"
)

const (
	electionTimeoutMin = 1 * time.Second
	electionTimeoutMax = 2 * time.Second
	heartbeatInterval  = 500 * time.Millisecond
)

func RandomElectionTimeout() time.Duration {
	delta := electionTimeoutMax - electionTimeoutMin
	return electionTimeoutMin + time.Duration(rand.Int63n(int64(delta)))
}