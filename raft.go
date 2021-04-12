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
	votes       map[string]struct{}
	votedFor    string
	net         *cnet.Network

	// leader info
	nextIndex  []int
	matchIndex []int
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
			fmt.Println("I am Leader", r.term)
			select {
			// receive client command
			// handle append responses
			case am := <-appendChan:
				r.handleAppendMsg(am, send)
			// handle election
			case lm := <-leaderChan:
				r.handleLeaderMsg(lm, send)
			// send regular updates faster than heartbeat timeout
			case <-time.After(100 * time.Millisecond):
				// Send empty Append
				r.sendAppendMsg(send)
			}
		case Follower:
			fmt.Println("I am Follower", r.term)
			select {
			// respond to append request
			case am := <-appendChan:
				r.handleAppendMsg(am, send)
			// response to leader requests
			case lm := <-leaderChan:
				r.handleLeaderMsg(lm, send)
			// elect a new leader
			case <-time.After(electionTimeout):
				r.becomeCandidate(send)
			}
		case Candidate:
			fmt.Println("I am Candidate", r.term)
			select {
			case am := <-appendChan:
				// Another is claiming leader
				r.handleAppendMsg(am, send)
			case lm := <-leaderChan:
				// Another candidate?
				// Response from voter?
				r.handleLeaderMsg(lm, send)
			case <-time.After(electionTimeout):
				r.becomeCandidate(send)
			}
		}
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

func (r *Raft) becomeLeader() {
	r.role = Leader

	r.nextIndex = make([]int, len(r.peers))
	r.matchIndex = make([]int, len(r.peers))
	for i := range r.peers {
		r.nextIndex[i] = len(r.log)
		r.matchIndex[i] = 0
	}

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
	r.votes = map[string]struct{}{r.peers[r.id]: {}}
	r.votedFor = r.peers[r.id]

	// Send request vote to all others
	for i, peer := range r.peers {
		if i == r.id {
			continue
		}

		lm := LeaderMsg{
			Term:         r.term,
			LastLogIndex: len(r.log) - 1,
			LastLogTerm:  r.logTerms[len(r.log)-1],
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

func (r *Raft) handleAppendMsg(am AppendMsg, send chan cnet.PeerMsg) {
	if am.Term < r.term {
		r.rejectAppendMsg(am, send)
		return
	} else if am.Term > r.term {
		r.becomeFollower(am.Term)
		r.handleAppendMsg(am, send)
		return
	} else if am.Response {
		r.handleAppendMsgResponse(am, send)
		return
	}

	// handle same term request
	// TODO for now do nothing
	am.Dst = am.Src
	am.Src = r.peers[r.id]
	am.Response = true
	am.Success = true

	amBytes, err := json.Marshal(am)
	if err != nil {
		fmt.Println(err)
		return
	}

	pm := cnet.PeerMsg{
		Src:  am.Src,
		Dst:  am.Dst,
		Msg:  amBytes,
		Type: APPEND,
	}

	send <- pm
}

func (r *Raft) handleAppendMsgResponse(am AppendMsg, send chan cnet.PeerMsg) {
	// TODO
}

func (r *Raft) rejectAppendMsg(am AppendMsg, send chan cnet.PeerMsg) {
	// respond to out-dated candidate
	am.Dst = am.Src
	am.Src = r.peers[r.id]
	am.Term = r.term
	am.Response = true
	am.Success = false

	amBytes, err := json.Marshal(am)
	if err != nil {
		fmt.Println(err)
		return
	}

	pm := cnet.PeerMsg{
		Src:  am.Src,
		Dst:  am.Dst,
		Msg:  amBytes,
		Type: APPEND,
	}

	send <- pm
}

func (r *Raft) sendAppendMsg(send chan cnet.PeerMsg) {
	for i, peer := range r.peers {
		if i == r.id {
			continue
		}

		am := AppendMsg{
			Src:          r.peers[r.id],
			Dst:          peer,
			Term:         r.term,
			PrevLogIndex: r.nextIndex[i] - 1,
			PrevLogTerm:  r.logTerms[r.nextIndex[i]-1],
			Entries:      r.log[r.nextIndex[i]:],
			LeaderCommit: r.commitIndex,
		}
		amBytes, err := json.Marshal(am)
		if err != nil {
			fmt.Println(err)
			continue
		}

		pm := cnet.PeerMsg{
			Src:  r.peers[r.id],
			Dst:  peer,
			Msg:  amBytes,
			Type: APPEND,
		}

		send <- pm
	}
}

func (r *Raft) handleLeaderMsgResponse(lm LeaderMsg) {
	if !lm.VoteGranted {
		return
	}

	r.votes[lm.Src] = struct{}{}

	if len(r.votes) > (len(r.peers) / 2) {
		r.becomeLeader()
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
	if lm.Term < r.term {
		r.rejectLeaderMsg(lm, send)
		return
	} else if lm.Term > r.term {
		r.becomeFollower(lm.Term)
		r.handleLeaderMsg(lm, send)
		return
	} else if lm.Response {
		r.handleLeaderMsgResponse(lm)
		return
	}

	lm.Dst = lm.Src
	lm.Src = r.peers[r.id]
	lm.Response = true

	var isUpToDate bool = false
	if lm.LastLogTerm > r.logTerms[len(r.log)-1] {
		isUpToDate = true
	} else if lm.LastLogTerm == r.logTerms[len(r.log)-1] &&
		lm.LastLogIndex >= len(r.log)-1 {
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
