package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/amitbet/teleporter/common"
	"github.com/amitbet/teleporter/logger"
	"github.com/inconshreveable/muxado"
)

type TunnelTask struct {
	Conn    net.Conn
	PreSend []byte // any bytes that need to be sent to the other side before piping the connections together
}

type Tunnel struct {
	session  muxado.Session
	taskChan chan *TunnelTask
	QuitChan chan bool
	// username string
	// password string
	Name string
}

func NewTunnel(sess muxado.Session, tasks chan *TunnelTask, name string) *Tunnel {
	tun := &Tunnel{
		session:  sess,
		taskChan: tasks,
		// username: user,
		// password: pass,
		Name:     name,
		QuitChan: make(chan bool),
	}
	return tun
}

func (t *Tunnel) Run() {
	for {
		select {
		case task := <-t.taskChan:
			//5s to get the socks establishing over with or bust
			////task.Conn.SetDeadline(time.Now().Add(5 * time.Second))

			go t.newSocksConn(task)
		case <-t.QuitChan:
			fmt.Println("tunnel: " + t.Name + " quitting!")
			return
		}
	}

}

type client struct {
	tunnels []*Tunnel
	queue   chan net.Conn // queue of connections for worker sessions to handle
	// username      string
	// password      string
	IncomingConns chan *TunnelTask
	Config        *common.ClientConfig
	Name          string
	//Remoteip   string
	//Port       string
	//listener   net.Listener  // ---- remove ----
}

// CanAccept checks if the given target address exists in the shared network ip/domains for this client
func (c *client) CanAccept(address string) bool {
	// dumb default implementation for now
	return true
}

func newClient(cfg *common.ClientConfig) *client {
	//_, p, _ := net.SplitHostPort(listener.Addr().String())
	//r, _, _ := net.SplitHostPort(session.Addr().String())
	c := &client{
		tunnels: []*Tunnel{},
		Config:  cfg,
		// username:      username,
		// password:      password,
		IncomingConns: make(chan *TunnelTask),
		Name:          cfg.ClientId,
		//Port:     p,
		//Remoteip: r,
	}
	//c.AddSession(session)
	return c
}

// AddTunnel adds a tunnel which manages a single connection to a client
func (c *client) AddTunnel(sess muxado.Session) {
	tun := NewTunnel(sess, c.IncomingConns, c.Name+strconv.Itoa(len(c.tunnels)))
	c.tunnels = append(c.tunnels, tun)
	go tun.Run()
}

// // listen manages the socks5 listener interface on the server
// func (c *client) AddTunnel(sess muxado.Session) {
// 	//defer c.listener.Close()
// 	for {
// 		// Accept a TCP connection
// 		// conn, err := c.listener.Accept()
// 		// if err != nil {
// 		// 	logger.Error("Closed tcp server: ", err)
// 		// 	return
// 		// }
// 		//5s to get the socks establishing over with or bust
// 		//conn.SetDeadline(time.Now().Add(5 * time.Second))

// 		//	go c.newSocksConn(conn)
// 	}
// }

func (c *client) wait(sess muxado.Session) {
	code, err, debug := sess.Wait() //wait for close
	logger.Info("Session mux shutdown with code %v error %v debug %v", code, err, debug)

	//we are done here, shut down the tcp listener
	//c.listener.Close()
}

func (t *Tunnel) newSocksConn(task *TunnelTask) {
	defer task.Conn.Close()

	logger.Debug("server: open mux connection back")
	//open mux connection back
	muxConn, err := t.session.Open()
	if err != nil {
		logger.Error("Error establishing session", err)
		return
	}
	defer muxConn.Close()

	//send "no auth" message & receive response
	muxConn.Write([]byte{socks5Version, 1, NoAuth})
	reply := []byte{0, 0}
	io.ReadAtLeast(muxConn, reply, 2)

	//remove deadline
	task.Conn.SetDeadline(time.Time{})

	//send the preSend = the bytes read to check dest-address in server
	muxConn.Write(task.PreSend)

	//From now on proxy it all
	errCh := make(chan error, 2)
	go proxy(task.Conn, muxConn, errCh)
	go proxy(muxConn, task.Conn, errCh)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			// return from this function closes target (and conn).
			return
		}
	}
}

type closeWriter interface {
	CloseWrite() error
}

// proxy is used to suffle data from src to destination, and sends errors
// down a dedicated channel
func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	q, err := io.Copy(dst, src)
	logger.Infof("transferred %d bytes", q)
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
	errCh <- err
}
