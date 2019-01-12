package client

import (
	"io"
	"math/rand"
	"net"
	"sync"

	"github.com/amitbet/teleporter/logger"
	"github.com/inconshreveable/muxado"
)

// MultiMux is a client for multiple mux channels,
// it presents a facade which makes multiple channels seem like one channel
type MultiMux struct {
	connections []muxado.Session
	sconns      chan net.Conn
	isClient    bool
	mut         sync.RWMutex
}

// NewMultiMux creates a new multi connection mux
func NewMultiMux(isClient bool) *MultiMux {
	mm := &MultiMux{}
	mm.sconns = make(chan net.Conn, 16)
	mm.isClient = isClient
	mm.mut = sync.RWMutex{}
	return mm
}

// AddConnection adds a connection to the multi-mux
func (m *MultiMux) AddConnection(c io.ReadWriteCloser) {
	var sess muxado.Session
	if m.isClient {
		sess = muxado.Client(c, nil)
	} else {
		sess = muxado.Server(c, nil)
	}

	m.mut.Lock()
	m.connections = append(m.connections, sess)
	m.mut.Unlock()

	go m.handleSession(sess)
}

func (m *MultiMux) handleSession(sess muxado.Session) {
	for {
		sconn, err := sess.Accept()
		if err != nil {
			logger.Error("Can't accept, connection is dead", err)
			break
		}
		m.sconns <- sconn
	}
}

// Accept returns an incoming client connection or waits until one is initiated
func (m *MultiMux) Accept() (net.Conn, error) {
	sconn := <-m.sconns

	return sconn, nil
}

func (m *MultiMux) Open() (net.Conn, error) {
	connCount := len(m.connections)
	pick := rand.Intn(connCount)

	m.mut.RLock()
	sess := m.connections[pick]
	m.mut.RUnlock()

	return sess.Open()
}
