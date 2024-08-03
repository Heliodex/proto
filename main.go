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

type AddrSet = map[string]*net.UDPAddr

var (
	local       = net.IPv6loopback
	laddr       *net.UDPAddr
	server      *net.UDPConn
	unconfirmed = make(AddrSet)
	confirmed   = make(AddrSet)
)

var gossip = []byte("Hello!") // broadcasted to confirmed nodes every second

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

func broadcast(msgType byte, msg []byte) {
	for _, host := range confirmed {
		fmt.Println(c.InPurple(fmt.Sprintf("  To node %s: %s", vis(host.Port), msg)))

		_, err := server.WriteToUDP(append([]byte{msgType}, msg...), host)
		if err != nil {
			panic(err)
		}
	}
}

func network() {
	for _, host := range unconfirmed {
		w := new(bytes.Buffer)
		enc := gob.NewEncoder(w)
		nodes := allNodes()
		delete(nodes, host.String())
		enc.Encode(nodes)

		fmt.Println(c.InPurple(fmt.Sprintf("  To node %s: You know these guys?", vis(host.Port))))

		_, err := server.WriteToUDP(append([]byte{1}, w.Bytes()...), host)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	if len(os.Args) < 1 {
		panic("needs arg")
	}

	var lport int
	fmt.Sscanf(os.Args[1], "%d", &lport)
	lport += 10000
	laddr = &net.UDPAddr{IP: local, Port: lport}

	var err error
	server, err = net.ListenUDP("udp", laddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("I am node", vis(lport))

	for _, v := range os.Args[2:] {
		var port int
		fmt.Sscanf(v, "%d", &port)

		addr := &net.UDPAddr{IP: local, Port: port + 10000}

		if addr.String() == laddr.String() {
			panic("given own port as arg")
		}
		unconfirmed[addr.String()] = addr
	}
	fmt.Println(len(unconfirmed), "unconfirmed nodes known")

	go func() {
		for {
			network()
			broadcast(0, gossip)
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		req := make([]byte, 1024)
		n, addr, _ := server.ReadFromUDP(req)
		reqType := req[0]
		req = req[1:n]

		switch reqType {
		case 1: // distribute
			fmt.Println(c.InBlue(fmt.Sprintf("From node %s: You know these guys?", vis(addr.Port))))

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

		default: // gossip
			fmt.Println(c.InGreen(fmt.Sprintf("From node %s: %s", vis(addr.Port), req)))
		}

		if addr.String() == laddr.String() || confirmed[addr.String()] != nil {
			continue
		}
		fmt.Println(c.InYellow("  Confirmed node " + vis(addr.Port)))
		delete(unconfirmed, addr.String())
		confirmed[addr.String()] = addr
	}
}
