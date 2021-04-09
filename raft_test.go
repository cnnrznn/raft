package raft

import (
	"testing"
	"time"
)

func TestHeartBeats(t *testing.T) {
	// start http server and read endpoints

	heartBeatChan := make(chan struct{})
	candidateChan := make(chan struct{})
	r := New(heartBeatChan, candidateChan)

	go r.Run()

	for i := 0; i < 10; i++ {
		time.Sleep(600 * time.Millisecond)
		heartBeatChan <- struct{}{}
	}
}
