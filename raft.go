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
	id          int
	peers       []string
	term        int
	log         []string
	logTerms    []int
	commitIndex int
	role        Role
	votes       []string
	votedFor    string
	net         *cnet.Network
}

func New(
	id int,
	peers []string,
) *Raft {
	return &Raft{
		id:          id,
		peers:       peers,
		role:        Follower,
		net:         cnet.New(id, peers),
		commitIndex: 0,
		log:         []string{""},
		logTerms:    []int{0},
		term:        0,
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
		electionTimeout := time.Duration(rand.Intn(500)+500) * time.Millisecond
		switch r.role {
		case Leader:
			select {
			// receive client command
			// handle leader requests
			case lm := <-leaderChan:
				r.asLeaderOrFollowerHandleLeaderMsg(lm, send)
			// send regular updates faster than heartbeat timeout
			case <-time.After(100 * time.Millisecond):
				// Send empty Append
			}
		case Follower:
			select {
			// respond to append request
			case <-appendChan:
			// response to leader requests
			case lm := <-leaderChan:
				r.asLeaderOrFollowerHandleLeaderMsg(lm, send)
			// elect a new leader
			case <-time.After(electionTimeout):
				r.becomeCandidate(send)
			}
		case Candidate:
			select {
			case <-appendChan:
				// Another is claiming leader
			case lm := <-leaderChan:
				// Another candidate?
				// Response from voter?
				r.asCandidateHandleLeaderMsg(lm, send)
			case <-time.After(electionTimeout):
				r.becomeCandidate(send)
			}
		}
	}
}

func (r *Raft) becomeLeader() {
	r.role = Leader

	// reset the vote
	r.votedFor = ""
}

func (r *Raft) becomeFollower(term int) {
	r.role = Follower
	r.term = term

	// reset the vote
	r.votedFor = ""
}

func (r *Raft) becomeCandidate(send chan cnet.PeerMsg) {
	r.role = Candidate
	r.term++
	r.votes = []string{r.peers[r.id]}
	r.votedFor = r.peers[r.id]

	// Send request vote to all others
	for i, peer := range r.peers {
		if i == r.id {
			continue
		}

		lm := LeaderMsg{
			Term:         r.term,
			LastLogIndex: r.commitIndex,
			LastLogTerm:  r.logTerms[r.commitIndex],
			Src:          r.peers[r.id],
			Dst:          peer,
		}
		lmBytes, err := json.Marshal(lm)
		if err != nil {
			fmt.Println(err)
			continue
		}

		pm := cnet.PeerMsg{
			Type: LEADER,
			Msg:  lmBytes,
			Src:  r.peers[r.id],
			Dst:  peer,
		}

		send <- pm
	}
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

func (r *Raft) asLeaderOrFollowerHandleLeaderMsg(
	lm LeaderMsg,
	send chan cnet.PeerMsg,
) {
	if lm.Response {
		// As a leader or follower ignore responses to old requests
		// I am no longer a candidate
		return
	}

	if lm.Term < r.term {
		r.rejectLeaderMsg(lm, send)
	} else if lm.Term > r.term {
		r.becomeFollower(lm.Term)
		r.handleLeaderMsg(lm, send)
	}
}

func (r *Raft) asCandidateHandleLeaderMsg(
	lm LeaderMsg,
	send chan cnet.PeerMsg,
) {
	if lm.Term < r.term {
		r.rejectLeaderMsg(lm, send)
	} else if lm.Term > r.term {
		r.becomeFollower(lm.Term)
		r.handleLeaderMsg(lm, send)
	}
}

func (r *Raft) rejectLeaderMsg(lm LeaderMsg, send chan cnet.PeerMsg) {
	// respond to out-dated candidate
	lm.Dst = lm.Src
	lm.Src = r.peers[r.id]
	lm.Term = r.term
	lm.Response = true
	lm.VoteGranted = false

	lmBytes, err := json.Marshal(lm)
	if err != nil {
		fmt.Println(err)
		return
	}

	pm := cnet.PeerMsg{
		Src:  lm.Src,
		Dst:  lm.Dst,
		Msg:  lmBytes,
		Type: LEADER,
	}

	send <- pm
}

func (r *Raft) handleLeaderMsg(lm LeaderMsg, send chan cnet.PeerMsg) {
	lm.Dst = lm.Src
	lm.Src = r.peers[r.id]
	lm.Response = true

	var isUpToDate bool = false
	if lm.LastLogTerm > r.logTerms[r.commitIndex] {
		isUpToDate = true
	} else if lm.LastLogTerm == r.logTerms[r.commitIndex] &&
		lm.LastLogIndex >= r.commitIndex {
		isUpToDate = true
	}

	if lm.Term < r.term {
		lm.VoteGranted = false
	} else if (r.votedFor == "" || r.votedFor == lm.Dst) &&
		isUpToDate {
		lm.VoteGranted = true
		r.votedFor = lm.Dst
	}

	lmBytes, err := json.Marshal(lm)
	if err != nil {
		fmt.Println(err)
		return
	}

	pm := cnet.PeerMsg{
		Src:  lm.Src,
		Dst:  lm.Dst,
		Msg:  lmBytes,
		Type: LEADER,
	}

	send <- pm
}
