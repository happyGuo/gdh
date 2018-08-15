package gdhd

import (
	"sync"
	"sync/atomic"
	"time"
	"net"
	"crypto/tls"
	"os"
	"log"

	"gdhMQ/internal/protocol"
	"gdhMQ/internal/util"
)

type GDHD struct {
	// 64bit atomic vars need to be first for proper alignment on 32bit platforms
	clientIDSequence int64

	sync.RWMutex

	opts atomic.Value

	isLoading int32
	errValue  atomic.Value
	startTime time.Time

	

	lookupPeers atomic.Value

	tcpListener   net.Listener
	httpListener  net.Listener
	httpsListener net.Listener
	tlsConfig     *tls.Config

	poolSize int

	notifyChan           chan interface{}
	optsNotificationChan chan struct{}
	exitChan             chan int
	waitGroup            util.WaitGroupWrapper

}

func (g *GDHD) Entry(){
	var err error
	addr := "0.0.0.0:8848"
	ctx := &context{g}
	g.tcpListener, err = net.Listen("tcp", addr,)
	if err != nil {
		log.Printf("listen (%s) failed - %s",addr,err)
		os.Exit(1)
	}
	tcpServer := &tcpServer{ctx: ctx}
	g.waitGroup.Wrap(func() {
		protocol.TCPServer(g.tcpListener, tcpServer)
	})

}

func New() *GDHD  {
	g := &GDHD{
		startTime:            time.Now(),
		exitChan:             make(chan int),
		notifyChan:           make(chan interface{}),
		optsNotificationChan: make(chan struct{}, 1),
	}

	return g
}
