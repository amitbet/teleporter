package main

import (
	"io"
	"net"
	"time"

	"github.com/amitbet/teleporter/logger"
	"github.com/inconshreveable/muxado"
)

type client struct {
	tunnel   muxado.Session
	listener net.Listener
	Username string
	Password string
	Remoteip string
	Port     string
}

func newClient(session muxado.Session, listener net.Listener, username string, password string) *client {
	_, p, _ := net.SplitHostPort(listener.Addr().String())
	r, _, _ := net.SplitHostPort(session.Addr().String())
	c := &client{
		tunnel:   session,
		listener: listener,
		Username: username,
		Password: password,
		Port:     p,
		Remoteip: r,
	}

	return c
}

func (c *client) listen() {
	//defer c.listener.Close()
	for {
		// Accept a TCP connection
		conn, err := c.listener.Accept()
		if err != nil {
			logger.Error("Closed tcp server: ", err)
			return
		}
		//5s to get the socks establishing over with or bust
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		go c.newSocksConn(conn)
	}
}

func (c *client) wait() {
	code, err, debug := c.tunnel.Wait() //wait for close
	logger.Info("Session mux shutdown with code %v error %v debug %v", code, err, debug)

	//we are done here, shut down the tcp listener
	//c.listener.Close()
}

func (c *client) newSocksConn(conn net.Conn) {
	defer conn.Close()

	//do auth
	if err := auth(conn, c.Username, c.Password); err != nil {
		logger.Error("auth problem: ", err)
		return
	}

	logger.Debug("server: open mux connection back")
	//open mux connection back
	tunnel, err := c.tunnel.Open()
	if err != nil {
		logger.Error("Error establishing session", err)
		return
	}
	defer tunnel.Close()

	//send "no auth" message & receive response
	tunnel.Write([]byte{socks5Version, 1, NoAuth})
	reply := []byte{0, 0}
	io.ReadAtLeast(tunnel, reply, 2)

	//remove deadline
	conn.SetDeadline(time.Time{})

	//From now on proxy it all
	errCh := make(chan error, 2)
	go proxy(conn, tunnel, errCh)
	go proxy(tunnel, conn, errCh)

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
