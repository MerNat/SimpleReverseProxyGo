package main

import (
	"flag"
	"log"
	"net"

	"github.com/MerNat/SimpleReverseProxyGoLang/caching"
	"github.com/MerNat/SimpleReverseProxyGoLang/proxy"
)

func main() {
	localAddr := flag.String("l", ":8080", "Local address")
	remoteAddr := flag.String("r", "localhost:3000", "Remote address")
	cacheExpiration := flag.Int("c", 4, "Cache Expiration in seconds")

	flag.Parse()

	caching.CacheExpiration = *cacheExpiration
	listenerAddress, err := caching.TCPAddressResolver(*localAddr)
	if err != nil {
		log.Fatalf("Failed to resolve local address: %v", err)
	}

	remoteAddress, err := caching.TCPAddressResolver(*remoteAddr)

	if err != nil {
		log.Fatalf("Failed to resolve remote address: %v", err)
	}

	listener, err := net.ListenTCP("tcp", listenerAddress)

	if err != nil {
		log.Fatalf("Failed to open local port to listen: %v", err)
	}

	log.Printf("Simple Proxy started on: %d and forwards to port %d: (Caching Expiration: %d Seconds)", listenerAddress.Port, remoteAddress.Port, caching.CacheExpiration)
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
		go p.Start()
	}
}
