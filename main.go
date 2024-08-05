package main

import (
	"bytes"
	"encoding/gob" // don't get used to it. here until we figure out a better address format
	"fmt"
	"net"
	"os"
	"time"

	c "github.com/TwiN/go-color"
)

type (
	Node struct {
		addr      *net.UDPAddr
		confirmed bool
	}
	AddrSet = map[string]*net.UDPAddr // hash map with addr.String() as the key (works??)
	NodeSet = map[string]*Node
	Server  struct {
		*net.UDPConn
	}
)

var (
	local      = net.IPv6loopback
	laddr      *net.UDPAddr
	knownNodes = make(NodeSet)
	types      = map[byte]string{
		'H': "Hello!",
		'D': "Discovery",
	}
	lport int
)

func vis(port int) string {
	return fmt.Sprintf("%d", port-10000)
}

func getNodes(confirmed bool) AddrSet {
	set := make(AddrSet)
	for i, v := range knownNodes {
		if v.confirmed == confirmed {
			set[i] = v.addr
		}
	}
	return set
}

func (s Server) send(node *net.UDPAddr, msgType byte, msg []byte) {
	fmt.Println(c.InPurple(fmt.Sprintf("  To node %s: %s", vis(node.Port), types[msgType])))
	notify(lport, "Send")

	s.WriteToUDP(append([]byte{msgType}, msg...), node) // did it send or not? idk, who cares, yolo, UDP BABYYYY
}

func (s Server) broadcast(msgType byte, msg []byte) {
	for _, node := range getNodes(true) {
		s.send(node, msgType, msg)
	}
}

func (s Server) network() {
	for _, node := range getNodes(false) {
		w := new(bytes.Buffer)
		enc := gob.NewEncoder(w)

		// copy knownNodes map
		nodes := make(NodeSet, len(knownNodes))
		for i, v := range knownNodes {
			if i != node.String() { // except the one we're sending to
				nodes[i] = v
			}
		}

		enc.Encode(nodes)
		s.send(node, 'D', w.Bytes())
	}
}

func SendLoop(s Server) {
	s.network()
	s.broadcast('H', []byte{})
	time.Sleep(1 * time.Second)
}

func ReceiveLoop(s Server) {
	req := make([]byte, 1024)
	n, addr, _ := s.ReadFromUDP(req)
	reqType, content := req[0], req[1:n]

	switch reqType {
	case 'H': // gossip
		fmt.Println(c.InGreen(fmt.Sprintf("From node %s: %s", vis(addr.Port), types['H'])))
		notify(lport, "Receive")
	case 'D': // discovery
		fmt.Println(c.InBlue(fmt.Sprintf("From node %s: %s", vis(addr.Port), types['D'])))
		notify(lport, types['D'])

		var newNodes AddrSet // all discovered nodes are unconfirmed at first
		w := bytes.NewReader(content)
		gob.NewDecoder(w).Decode(&newNodes)

		for i, v := range newNodes {
			if v.String() == laddr.String() || knownNodes[i] != nil {
				continue
			}
			knownNodes[i] = &Node{v, false}
			fmt.Println(c.InYellow("Unconfirmed node " + vis(v.Port)))
		}
		s.network()
	default: // unknown
		fmt.Println(c.InGreen(fmt.Sprintf("From node %s: Unknown message", vis(addr.Port))))
		notify(lport, "Unknown message")
	}

	if addr.String() == laddr.String() || getNodes(true)[addr.String()] != nil {
		return
	}
	fmt.Println(c.InYellow("  Confirmed node " + vis(addr.Port))) // log here beacuse makes sense
	notify(lport, "Confirmed")

	// tell other nodes about new confirmed node...
	for _, node := range getNodes(true) {
		w := new(bytes.Buffer)
		gob.NewEncoder(w).Encode(AddrSet{addr.String(): addr})

		s.send(node, 'D', w.Bytes())
	}

	// ...BEFORE adding it to our confirmed nodes, to avoid infinite network loops (the worst kind of loops)
	knownNodes[addr.String()] = &Node{addr, true}
}

func main() {
	if len(os.Args) < 1 {
		fmt.Println(c.InRed("Too few arguments provided"))
		os.Exit(2)
	}

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
		knownNodes[addr.String()] = &Node{addr, false}
	}
	msg := fmt.Sprintf("%d unconfirmed nodes known", len(getNodes(false)))
	fmt.Println(msg)
	notify(lport, msg)

	go func() {
		for {
			SendLoop(s)
		}
	}()

	for {
		ReceiveLoop(s)
	}
}
