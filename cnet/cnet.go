package cnet

import (
	"encoding/json"
	"fmt"
	"net"
)

type PeerMsg struct {
	Src, Dst string
	Msg      string
}

type Network struct {
	id    int
	peers []string
}

func New(
	id int,
	peers []string,
) *Network {
	return &Network{
		id:    id,
		peers: peers,
	}
}

func (n *Network) Run(send, recv chan PeerMsg) {
	connChan := make(chan net.Conn)
	go n.listen(connChan)

	for {
		select {
		case ch := <-connChan:
			var pm PeerMsg
			bytes := make([]byte, 1024)
			n, err := ch.Read(bytes)
			if err != nil {
				fmt.Println(err)
			}
			err = json.Unmarshal(bytes[:n], &pm)
			if err != nil {
				fmt.Println(err)
			}

			recv <- pm
		case pm := <-send:
			go n.sendMsg(pm)
		}
	}
}

func (n *Network) sendMsg(pm PeerMsg) {
	conn, err := net.Dial("tcp", pm.Dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	bytes, err := json.Marshal(pm)
	if err != nil {
		fmt.Println(err)
	}
	_, err = conn.Write(bytes)
	if err != nil {
		fmt.Println(err)
	}
}

func (n *Network) listen(ch chan net.Conn) {
	ls, err := net.Listen("tcp", n.peers[n.id])
	if err != nil {
		return
	}

	for {
		conn, err := ls.Accept()
		if err != nil {
			continue
		}
		ch <- conn
	}
}
