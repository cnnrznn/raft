package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/cnnrznn/raft"
)

type config struct {
	Peers []string `json:"peers"`
	Self  int      `json:"self"`
}

func main() {
	// Read peer file and my address
	config := readPeers()
	args := os.Args
	id, _ := strconv.Atoi(args[1])
	config.Self = id

	// Make raft instance
	raft := raft.New(config.Self, config.Peers)
	fmt.Println(raft)

	// Run raft
	raft.Run()

	// Open http endpoint and take input from users
	// Allow the user to query for log entries
}

func readPeers() *config {
	bytes, err := os.ReadFile("peers.json")
	if err != nil {
		fmt.Errorf("Error reading peer file")
		return nil
	}

	var result config
	json.Unmarshal(bytes, &result)

	return &result
}
