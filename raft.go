package raft

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cnnrznn/raft/cnet"
)

type Role string

const (
	Leader    Role = "leader"
	Follower  Role = "follower"
	Candidate Role = "candidate"
)

type Raft struct {
	id            int
	peers         []string
	term          int
	log           []string
	role          Role
	votes         []string
	heartBeatChan <-chan struct{}
	candidateChan <-chan struct{}
	net           *cnet.Network
}

func New(
	id int,
	peers []string,
) *Raft {
	return &Raft{
		id:    id,
		peers: peers,
		role:  Follower,
		net:   cnet.New(id, peers),
	}
}

func (r *Raft) Run() {
	send := make(chan cnet.PeerMsg, 100)
	recv := make(chan cnet.PeerMsg, 100)
	go r.net.Run(send, recv)

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

func (r *Raft) becomeLeader() {
	r.role = Leader
}

func (r *Raft) becomeFollower() {
	r.role = Follower
}

func (r *Raft) becomeCandidate() {
	r.role = Candidate
	r.term++
	r.votes = []string{r.peers[r.id]}
	// Send request vote to all others
}
