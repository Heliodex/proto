package main

import (
	"fmt"
	"net"
	"os"
	"time"

	c "github.com/TwiN/go-color"
)

var (
	local      = net.IPv6loopback
	laddr      *net.UDPAddr
	knownNodes []*net.UDPAddr
	server     *net.UDPConn
)

const (
	gossip     = "Hello!" // broadcasted every second
	distribute = "You know these guys?" // broadcasted to every new node
)

func vis(port int) string {
	return fmt.Sprintf("%d", port-10000)
}

func in(arr []*net.UDPAddr, addr *net.UDPAddr) bool {
	for _, v := range arr {
		if (v.IP.Equal(addr.IP) && v.Port == addr.Port) &&
			!(v.IP.Equal(laddr.IP) && v.Port == laddr.Port) { // can't register self
			return true
		}
	}
	return false
}

func broadcast(msg string) {
	for _, host := range knownNodes {
		fmt.Println(c.InPurple(fmt.Sprintf("  To node %s: %s", vis(host.Port), msg)))

		_, err := server.WriteToUDP([]byte(msg), host)
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
		knownNodes = append(knownNodes, &net.UDPAddr{IP: local, Port: port + 10000})
	}
	fmt.Println(len(knownNodes), "nodes known")

	broadcast(distribute)

	go func() {
		for {
			broadcast(gossip)
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		req := make([]byte, 1024)
		n, addr, _ := server.ReadFromUDP(req)
		req = req[:n]

		if !in(knownNodes, addr) {
			fmt.Println(c.InYellow("Discovered node " + vis(addr.Port)))
			knownNodes = append(knownNodes, addr)
		}

		fmt.Println(c.InGreen(fmt.Sprintf("From node %s: %s", vis(addr.Port), req)))
	}
}
