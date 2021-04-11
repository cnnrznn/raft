package raft

type AppendMsg struct {
}

type LeaderMsg struct {
	// Request values
	Term         int `json:"term"`
	LastLogIndex int `json:"last_log_index"`
	LastLogTerm  int `json:"last_log_term"`

	// Response values
	// Term
	VoteGranted bool `json:"vote_granted"`
}
