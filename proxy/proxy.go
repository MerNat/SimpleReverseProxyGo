package proxy

import (
	"io"
	"log"
	"net"

	"github.com/MerNat/SimpleReverseProxyGoLang/util"
)

//Proxy identifies proxy as as a one way connection
type Proxy struct {
	Laddr, Raddr *net.TCPAddr
	Lconn, Rconn io.ReadWriteCloser
	ErrorSignal  chan bool
}

// NewConnection Create a new Proxy instance.
func NewConnection(lconn *net.TCPConn, laddr, raddr *net.TCPAddr) *Proxy {
	return &Proxy{
		Lconn:       lconn,
		Laddr:       laddr,
		Raddr:       raddr,
		ErrorSignal: make(chan bool),
	}
}

//Start initiates transmission of data to and from the remote to client side.
func (proxy *Proxy) Start() {
	defer proxy.Lconn.Close()

	var err error

	proxy.Rconn, err = net.DialTCP("tcp", nil, proxy.Raddr)

	if err != nil {
		log.Fatalf("Remote connection failure: %v", err)
	}

	sync := make(chan int)

	defer proxy.Rconn.Close()
	go proxy.CopySrcDst(proxy.Lconn, proxy.Rconn, true, sync)
	go proxy.CopySrcDst(proxy.Rconn, proxy.Lconn, false, sync)

	//Wait for everything to close -- This one blocks the routine.
	<-proxy.ErrorSignal
	log.Printf("Closing Start routine \n")
}

func (proxy *Proxy) err(err error) {
	if err != io.EOF {
		log.Printf("Warning: %v: Setting error signal to true", err)
	}
	proxy.ErrorSignal <- true
}

//CopySrcDst copies data from src to dest
func (proxy *Proxy) CopySrcDst(src, dst io.ReadWriteCloser, isFromLocalhost bool, id chan int) {
	buff := make([]byte, 6000)
	for {
		n, err := src.Read(buff)
		if err != nil {
			// Reading error.
			proxy.err(err)
			return
		}

		dataFromBuffer := buff[:n]

		if isFromLocalhost {
			identifier := len(util.Cache)
			cacheData, err := util.ExtractData(dataFromBuffer, identifier)

			checkOld, err := util.GetCacheDataUsingURL(cacheData.URL)
			if err != nil {
				cacheData = checkOld
			}
			index := cacheData.DoesCacheDataExistNB()
			if index >= 0 {
				//If found; retrieve data from cache and send it back to client.
				cData, err := util.GetCacheData(index)
				if err != nil {
					log.Println(err.Error())
				}
				log.Printf("[%s] Responding data from caching: \n", cData.URL)
				err = proxy.writeToDestination(src, cData.ResponseBody)
				if err != nil {
					proxy.err(err)
					// return
				}
				continue
			}
			// Save data for a later use.
			cacheData.AddCacheData()
			go func(ident int) {
				id <- ident
			}(identifier)
		} else {
			cacheData := &util.CacheData{ID: <-id}
			cacheData.SaveData(dataFromBuffer)
		}

		err = proxy.writeToDestination(dst, dataFromBuffer)
		if err != nil {
			proxy.err(err)
			return
		}
	}
}

func (proxy *Proxy) writeToDestination(destination io.ReadWriteCloser, data []byte) (err error) {
	_, err = destination.Write(data)
	return
}
