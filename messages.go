package raft

type AppendMsg struct {
	Response bool

	Src, Dst string

	// Request values
	Term         int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int

	// Response values
	// Term
	Success bool
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
