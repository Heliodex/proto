package main

import (
	"fmt"
	"net"
	"os"
	"time"

	c "github.com/TwiN/go-color"
)

var local = net.IPv6loopback

func main() {
	if len(os.Args) < 2 {
		panic("needs arg")
	}

	var lport int
	fmt.Sscanf(os.Args[1], "%d", &lport)
	laddr := &net.UDPAddr{IP: local, Port: lport}

	var rport int
	fmt.Sscanf(os.Args[2], "%d", &rport)
	raddr := &net.UDPAddr{IP: local, Port: rport}

	server, err := net.ListenUDP("udp", laddr)
	if err != nil {
		panic(err)
	}
	fmt.Println(laddr)
	fmt.Println(raddr)

	go func() {
		msg := fmt.Sprintf("Hello from %d!", lport)

		for {
			fmt.Println(c.InPurple("-> " + msg))
			_, err := server.WriteToUDP([]byte(msg), raddr)
			if err != nil {
				panic(err)
			}

			time.Sleep(2 * time.Second)
		}
	}()

	for {
		req := make([]byte, 1024)
		server.ReadFromUDP(req)

		fmt.Println(c.InGreen("<- " + string(req)))
	}
}
