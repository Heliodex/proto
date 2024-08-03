package main

import (
	"bytes"
	"encoding/gob" // don't get used to it.
	"fmt"
	"net"
	"os"
	"time"

	c "github.com/TwiN/go-color"
)

type (
	AddrSet = map[string]*net.UDPAddr
	Server  struct {
		*net.UDPConn
	}
)

var (
	local       = net.IPv6loopback
	laddr       *net.UDPAddr
	unconfirmed = make(AddrSet)
	confirmed   = make(AddrSet)
	types       = map[byte]string{
		'H': "Hello!",
		'D': "Discovery",
	}
)

func vis(port int) string {
	return fmt.Sprintf("%d", port-10000)
}

func allNodes() AddrSet {
	all := make(AddrSet)
	for i, v := range unconfirmed {
		all[i] = v
	}
	for i, v := range confirmed {
		all[i] = v
	}
	return all
}

func (s Server) send(node *net.UDPAddr, msgType byte, msg []byte) {
	fmt.Println(c.InPurple(fmt.Sprintf("  To node %s: %s", vis(node.Port), types[msgType])))
	s.WriteToUDP(append([]byte{msgType}, msg...), node) // did it send or not? idk, yolo, UDP BABYYYY
}

func (s Server) broadcast(msgType byte, msg []byte) {
	for _, node := range confirmed {
		s.send(node, msgType, msg)
	}
}

func (s Server) network() {
	for _, node := range unconfirmed {
		w := new(bytes.Buffer)
		enc := gob.NewEncoder(w)
		nodes := allNodes()
		delete(nodes, node.String()) // remove the other
		enc.Encode(nodes)

		s.send(node, 'D', w.Bytes())
	}
}

func main() {
	if len(os.Args) < 1 {
		fmt.Println(c.InRed("Too few arguments provided"))
		os.Exit(2)
	}

	var lport int
	fmt.Sscanf(os.Args[1], "%d", &lport)
	lport += 10000
	laddr = &net.UDPAddr{IP: local, Port: lport}

	server, err := net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Println(c.InRed("Error while starting server:"), err)
		os.Exit(1)
	}
	s := Server{server}

	fmt.Println("I am node", vis(lport))

	for _, v := range os.Args[2:] {
		var port int
		fmt.Sscanf(v, "%d", &port)

		addr := &net.UDPAddr{IP: local, Port: port + 10000}

		if addr.String() == laddr.String() {
			fmt.Println(c.InRed("Server's own address was given as a starting address"))
			os.Exit(2)
		}
		unconfirmed[addr.String()] = addr
	}
	fmt.Println(len(unconfirmed), "unconfirmed nodes known")

	go func() {
		for {
			s.network()
			s.broadcast('H', []byte{})
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		req := make([]byte, 1024)
		n, addr, _ := s.ReadFromUDP(req)
		reqType := req[0]
		req = req[1:n]

		switch reqType {
		case 'H': // gossip
			fmt.Println(c.InGreen(fmt.Sprintf("From node %s: %s", vis(addr.Port), types['H'])))
		case 'D': // discovery
			fmt.Println(c.InBlue(fmt.Sprintf("From node %s: %s", vis(addr.Port), types['D'])))

			var newUnconfirmed AddrSet
			w := bytes.NewReader(req)
			dec := gob.NewDecoder(w)
			dec.Decode(&newUnconfirmed)

			for i, v := range newUnconfirmed {
				if confirmed[i].String() == laddr.String() || unconfirmed[i] != nil || confirmed[i] != nil {
					continue
				}
				unconfirmed[i] = v
				fmt.Println(c.InYellow("Unconfirmed node " + vis(v.Port)))
			}
			s.network()
		default: // unknown
			fmt.Println(c.InGreen(fmt.Sprintf("From node %s: Unknown message", vis(addr.Port))))
		}

		if addr.String() == laddr.String() || confirmed[addr.String()] != nil {
			continue
		}
		fmt.Println(c.InYellow("  Confirmed node " + vis(addr.Port)))
		delete(unconfirmed, addr.String())
		confirmed[addr.String()] = addr
	}
}
