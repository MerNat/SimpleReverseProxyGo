package main

import (
	"log"
	"net"

	"github.com/MerNat/SimpleReverseProxyGoLang/proxy"
	"github.com/MerNat/SimpleReverseProxyGoLang/util"
)

func main() {
	util.CacheExpiration = 4
	listenerAddress, err := util.TCPAddressResolver(":8080")
	if err != nil {
		log.Fatalf("Failed to resolve local address: %v", err)
	}

	remoteAddress, err := util.TCPAddressResolver(":3000")

	if err != nil {
		log.Fatalf("Failed to resolve remote address: %v", err)
	}

	listener, err := net.ListenTCP("tcp", listenerAddress)

	if err != nil {
		log.Fatalf("Failed to open local port to listen: %v", err)
	}

	log.Printf("Simple Proxy started on: %d and forwards to port %d", listenerAddress.Port, remoteAddress.Port)
	for {
		conn, err := listener.AcceptTCP()

		if err != nil {
			log.Fatalf("Failed to accept connection: %v", err)
			continue
		}

		var p *proxy.Proxy
		// Http is a stateless protocol thus a proxy needes to reinitiate the new next incoming call (conn)
		// each time it finishes handling the previous one.
		p = proxy.NewConnection(conn, listenerAddress, remoteAddress)
		p.Start()
	}
}
