package raft

import (
	"github.com/google/uuid"
)

func (r *Raft) Request(msg string) bool {
	r.input <- Entry{
		Msg: msg,
		Id:  uuid.New(),
	}

	return <-r.success
}

func (r *Raft) Retrieve(start int) []Entry {
	start = max(start, 1)
	return r.log[start : r.commitIndex+1]
}
