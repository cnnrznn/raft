package raft

type AppendMsg struct {
	Response bool

	Term int
}

type LeaderMsg struct {
	// Request or response?
	Response bool

	Src, Dst string

	// Request values
	Term         int
	LastLogIndex int
	LastLogTerm  int

	// Response values
	// Term
	VoteGranted bool
}
