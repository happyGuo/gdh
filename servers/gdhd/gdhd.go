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
	"strings"
)

type GDHD struct {
	// 64bit atomic vars need to be first for proper alignment on 32bit platforms
	clientIDSequence int64

	sync.RWMutex

	opts atomic.Value

	isLoading int32
	errValue  atomic.Value
	startTime time.Time

	topicMap map[string]*Topic


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

// GetTopic performs a thread safe operation
// to return a pointer to a Topic object (potentially new)
func (n *GDHD) GetTopic(topicName string) *Topic {
	// most likely, we already have this topic, so try read lock first.
	n.RLock()
	t, ok := n.topicMap[topicName]
	n.RUnlock()
	if ok {
		return t
	}

	n.Lock()

	t, ok = n.topicMap[topicName]
	if ok {
		n.Unlock()
		return t
	}

	t = NewTopic(topicName, &context{n})
	n.topicMap[topicName] = t

	n.Unlock()

	// topic is created but messagePump not yet started

	// if loading metadata at startup, no lookupd connections yet, topic started after load
	if atomic.LoadInt32(&n.isLoading) == 1 {
		return t
	}

	// now that all channels are added, start topic messagePump
	t.Start()
	return t
}


func (n *GDHD) Notify(v interface{}) {
	// since the in-memory metadata is incomplete,
	// should not persist metadata while loading it.
	// nsqd will call `PersistMetadata` it after loading
	persist := atomic.LoadInt32(&n.isLoading) == 0
	n.waitGroup.Wrap(func() {
		// by selecting on exitChan we guarantee that
		// we do not block exit, see issue #123
		select {
		case <-n.exitChan:
		case n.notifyChan <- v:
			if !persist {
				return
			}

		}
	})
}
