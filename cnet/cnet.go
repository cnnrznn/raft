package cnet

import (
	"encoding/json"
	"fmt"
	"net"
)

type MessageType int

type PeerMsg struct {
	Src, Dst string
	Msg      []byte
	Type     MessageType
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
	connChan := make(chan net.Conn, 100)
	go n.listen(connChan)

	for {
		select {
		case conn := <-connChan:
			pm, err := recvMsg(conn)
			conn.Close()
			if err != nil {
				fmt.Println(err)
				continue
			}
			recv <- *pm
		case pm := <-send:
			n.sendMsg(pm)
		}
	}
}

func recvMsg(conn net.Conn) (*PeerMsg, error) {
	var pm PeerMsg
	bytes := make([]byte, 1024)

	n, err := conn.Read(bytes)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("reading from connection")
	}
	err = json.Unmarshal(bytes[:n], &pm)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("unmarshalling read data")
	}

	return &pm, nil
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
		return
	}
	_, err = conn.Write(bytes)
	if err != nil {
		fmt.Println(err)
	}
}

func (n *Network) listen(ch chan net.Conn) {
	ls, err := net.Listen("tcp", n.peers[n.id])
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := ls.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		ch <- conn
	}
}
