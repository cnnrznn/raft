package raft

import (
	"fmt"
	"math/rand"
	"time"
)

type Role string

const (
	Leader    Role = "leader"
	Follower  Role = "follower"
	Candidate Role = "candidate"
)

type Raft struct {
	name          string
	term          int
	log           []string
	role          Role
	votes         []string
	heartBeatChan <-chan struct{}
	candidateChan <-chan struct{}
}

func New(
	hbChan <-chan struct{},
	cdChan <-chan struct{},
) *Raft {
	return &Raft{
		role:          Follower,
		heartBeatChan: hbChan,
		candidateChan: cdChan,
	}
}

func (r *Raft) Run() {
	for {
		switch r.role {
		case Leader:
			select {
			// receive client command
			// send regular updates faster than heartbeat timeout
			case <-time.After(100 * time.Millisecond):
				// Send empty Append
			}
		case Follower:
			timeout := time.Duration(rand.Intn(500)+500) * time.Millisecond
			select {
			case <-r.heartBeatChan:
				fmt.Println("heart beat")
				// respond to heartbeat (leader request)
				// or leader
			case <-r.candidateChan:
				// respond to candidate request
			case <-time.After(timeout):
				r.becomeCandidate()
			}
		case Candidate:
			select {
			case <-r.heartBeatChan:
				// Another is claiming leader
			case <-r.candidateChan:
				// Another candidate?
				// Response from voter?
			}
		}
	}
}

func (r *Raft) becomeCandidate() {
	r.role = Candidate
	r.term++
	r.votes = []string{r.name}
	// Send request vote to all others
}
