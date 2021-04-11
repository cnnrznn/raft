package raft

import (
	"encoding/json"
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

const (
	APPEND cnet.MessageType = iota
	LEADER
)

type Raft struct {
	id       int
	peers    []string
	term     int
	log      []string
	role     Role
	votes    []string
	votedFor string
	net      *cnet.Network
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

	appendChan := make(chan AppendMsg, 100)
	leaderChan := make(chan LeaderMsg, 100)
	go route(recv, appendChan, leaderChan)

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
			case <-appendChan:
				// respond to heartbeat (leader request)
				// or leader
			case <-leaderChan:
				// respond to candidate request
			case <-time.After(timeout):
				r.becomeCandidate()
			}
		case Candidate:
			select {
			case <-appendChan:
				// Another is claiming leader
			case <-leaderChan:
				// Another candidate?
				// Response from voter?
			}
		}
	}
}

func (r *Raft) becomeLeader() {
	r.role = Leader

	// reset the vote
	r.votedFor = ""
}

func (r *Raft) becomeFollower() {
	r.role = Follower

	// reset the vote
	r.votedFor = ""
}

func (r *Raft) becomeCandidate() {
	r.role = Candidate
	r.term++
	r.votes = []string{r.peers[r.id]}
	r.votedFor = r.peers[r.id]
	// Send request vote to all others
}

func route(recv chan cnet.PeerMsg, appendChan chan AppendMsg, leaderChan chan LeaderMsg) {
	for {
		// Read messages from recv
		pm := <-recv

		// Parse payload
		// Forward message to correct channel
		switch pm.Type {
		case LEADER:
			var lm LeaderMsg
			err := json.Unmarshal(pm.Msg, &lm)
			if err != nil {
				fmt.Println(err)
				continue
			}
			leaderChan <- lm
		case APPEND:
			var am AppendMsg
			err := json.Unmarshal(pm.Msg, &am)
			if err != nil {
				fmt.Println(err)
				continue
			}
			appendChan <- am
		}
	}
}
