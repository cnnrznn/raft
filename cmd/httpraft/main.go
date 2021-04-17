package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cnnrznn/raft"
)

type config struct {
	Peers []string `json:"peers"`
	APIs  []string `json:"apis"`
}

type Server struct {
	raft   *raft.Raft
	config *config
}

func main() {
	// Read peer file and my address
	config := readPeers()
	args := os.Args
	id, _ := strconv.Atoi(args[1])

	// Make raft instance
	raft := raft.New(id, config.Peers)
	fmt.Println(raft)

	// Run raft
	go raft.Run()

	// continuously print commited log
	go func() {
		for {
			time.Sleep(5 * time.Second)
			log := []string{}
			entries := raft.Retrieve(0)
			for _, entry := range entries {
				log = append(log, entry.Msg)
			}
			fmt.Println(log)
		}
	}()

	server := Server{
		raft:   raft,
		config: config,
	}

	// Open http endpoint and take input from users
	// POST /request
	http.HandleFunc("/request", server.handleRequest)

	// Allow the user to query for log entries
	// GET /log?index=<int>
	// returns log starting at the provided index
	http.HandleFunc("/log", server.handleLog)

	log.Fatal(http.ListenAndServe(config.APIs[id], nil))
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

type Request struct {
	Entry string `json:"entry"`
}

type Response struct {
	Success bool   `json:"success"`
	Leader  string `json:"leader"`
}

type LogRequest struct {
	Index int `json:"index"`
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Must POST", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Problem parsing", http.StatusInternalServerError)
		return
	}

	// Submit to raft
	result := s.raft.Request(req.Entry)
	resp := Response{
		Success: result.Success,
		Leader:  s.config.APIs[result.Leader],
	}

	// Write response
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
}
