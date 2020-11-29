package main

import (
	"io"
	"log"
	"net"
)

//Proxy struct
type Proxy struct {
	laddr, raddr *net.TCPAddr
	lconn, rconn io.ReadWriteCloser
	errorSignal  chan bool
}

// New Create a new Proxy instance.
func New(lconn *net.TCPConn, laddr, raddr *net.TCPAddr) *Proxy {
	return &Proxy{
		lconn:       lconn,
		laddr:       laddr,
		raddr:       raddr,
		errorSignal: make(chan bool),
	}
}

//TCPAddressResolver resolved an address and returns to an struct having ip and port.
func TCPAddressResolver(addr string) (tcpAddress *net.TCPAddr, err error) {
	tcpAddress, err = net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return
}
func main() {
	listenerAddress, err := TCPAddressResolver(":8080")
	if err != nil {
		log.Fatalf("Failed to resolve local address: %v", err)
	}

	remoteAddress, err := TCPAddressResolver("localhost:3000")

	if err != nil {
		log.Fatalf("Failed to resolve remote address: %v", err)
	}

	listener, err := net.ListenTCP("tcp", listenerAddress)

	if err != nil {
		log.Fatalf("Failed to open local port to listen: %v", err)
	}
	for {
		conn, err := listener.AcceptTCP()

		if err != nil {
			log.Fatalf("Failed to accept connection: %v", err)
			continue
		}

		var p *Proxy
		// Http is a stateless protocol thus a proxy needes to reinitiate the new next incoming call (conn)
		// each time it finishes handling the previous one.
		p = New(conn, listenerAddress, remoteAddress)
		p.Start()
	}
}

//Start initiates transmission of data to and from the remote to client side.
func (p *Proxy) Start() {
	defer p.lconn.Close()

	var err error

	p.rconn, err = net.DialTCP("tcp", nil, p.raddr)

	if err != nil {
		log.Fatalf("Remote connection failure: %v", err)
	}

	defer p.rconn.Close()

	go p.CopySrcDst(p.lconn, p.rconn)
	go p.CopySrcDst(p.rconn, p.lconn)

	//Wait for everything to close -- This one blocks the routine.
	<-p.errorSignal
	log.Printf("Closing Start routine \n")
}

func (p *Proxy) err(err error) {
	if err != io.EOF {
		log.Printf("Warning: %v: Setting error signal to true", err)
	}
	p.errorSignal <- true
}

//CopySrcDst copies data from src to dest
func (p *Proxy) CopySrcDst(src, dst io.ReadWriteCloser) {
	buff := make([]byte, 1024)
	for {
		n, err := src.Read(buff)
		if err != nil {
			// Reading error.
			p.err(err)
			return
		}

		dataFromBuffer := buff[:n]

		n, err = dst.Write(dataFromBuffer)
		if err != nil {
			// Writing error.
			p.err(err)
			return
		}
	}
}
