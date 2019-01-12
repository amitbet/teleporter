package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/amitbet/go-socks5"
	"github.com/amitbet/teleporter/common"
	"github.com/amitbet/teleporter/logger"
	"github.com/inconshreveable/muxado"
	"github.com/pions/dtls/pkg/dtls"
)

var cfg *config

var clientjson []byte

var clients = make(map[string]*client)

const banner = `Teleproxy node!`

func createControlListener(listenerType string) (net.Listener, error) {
	var controlListener net.Listener
	var err error
	controlAddr := ":" + cfg.CnCPort

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return nil, err
	}
	tlsconfig := &tls.Config{Certificates: []tls.Certificate{cer}}

	switch listenerType {
	case "tls":
		controlListener, err = tls.Listen("tcp", controlAddr, tlsconfig)

		//l, err := net.Listen("tcp", ":"+cfg.CnCPort)

		if err != nil {
			return nil, err
		}

		logger.Debug("Started CnC server at port " + cfg.CnCPort)
	case "dtls":
		logger.Debug("running on dtls")
		addr, err := net.ResolveUDPAddr("udp", controlAddr)
		if err != nil {
			return nil, err
		}

		certificate, privateKey, err := dtls.GenerateSelfSigned()
		if err != nil {
			return nil, err
		}
		// Prepare the configuration of the DTLS connection
		config := &dtls.Config{Certificate: certificate, PrivateKey: privateKey}

		// Connect to a DTLS server
		controlListener, err = dtls.Listen("udp", addr, config)

		if err != nil {
			return nil, err
		}
	}
	return controlListener, nil
}

func createSocks5Listener(socksAddr string) (net.Listener, error) {

	socksListener, err := net.Listen("tcp", socksAddr)
	if err != nil {
		logger.Error("Error starting Socks listener", err)
		return nil, err
	}
	logger.Infof("Started new SOCKS listener at port %v\n", socksListener.Addr().String())
	return socksListener, nil
}

// handles an incomming socks connection
func handleSocks(listener net.Listener) {
	for {
		// Accept a TCP connection
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Closed tcp server: ", err)
			continue
		}
		// 5s to get the socks establishing over with
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		// connect & authenticate
		cator := socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				cfg.SocksUsername: cfg.SocksPass,
			},
		}
		req, err := socks5.PerformHandshake(conn, []socks5.Authenticator{cator})

		if err != nil {
			logger.Error("Error in socks5 handshake: ", err)
			continue
		}

		address := req.DestAddr.Address()
		for _, cl := range clients {
			if cl.CanAccept(address) {

				// adding the request to this task, so it can be handled on the other side -----
				buff := bytes.Buffer{}
				req.WriteTo(&buff)

				task := TunnelTask{
					Conn:    conn,
					PreSend: buff.Bytes(),
				}
				//logger.Debug("PreSend: ", task.PreSend)
				cl.IncomingConns <- &task
				break
			}
		}

		//	go c.newSocksConn(conn)
	}
}

//Listen create a listener and serve on it
func listen(listenerType string) error {
	//incomingChan := make(chan net.Conn)

	//------------ socks 5 server --------------
	socksAddr := ":" + cfg.SocksPort
	socksListener, err := createSocks5Listener(socksAddr)
	if err != nil {
		return err
	}
	defer socksListener.Close()
	go handleSocks(socksListener)

	//------------ tunneler server --------------
	controlListener, err := createControlListener(listenerType)
	if err != nil {
		return err
	}
	defer controlListener.Close()
	for {
		conn, err := controlListener.Accept()
		if err != nil {
			logger.Error("TCP accept failed: %s\n", err)
			continue
		}

		go handleClientConn(conn)
	}
}

// handle manages a new physical client connection comming into the control port
func handleClientConn(conn net.Conn) {

	// read client configuration
	clientConfigStr, err := common.ReadString(conn)
	if err != nil {
		logger.Error("Client connect, failed while reading client header: %s\n", err)
		return
	}

	logger.Debug("client connected, read client config string: ", clientConfigStr)
	cconfig := common.ClientConfig{}
	err = json.Unmarshal([]byte(clientConfigStr), &cconfig)
	if err != nil {
		logger.Error("Client connect, error unmarshaling clientConfig: %s\n", err)
		return
	}
	cid := cconfig.ClientId
	logger.Info("Client connected, id: ", cid)

	// Setup server side of muxado
	session := muxado.Server(conn, nil)
	defer session.Close()

	var theClient *client
	var ok bool
	theClient, ok = clients[cid]
	if !ok {
		//create new Client
		theClient = newClient(&cconfig)
		clients[cid] = theClient
		theClient.AddTunnel(session)
		//go theClient.listen()
	} else {
		theClient.AddTunnel(session)
	}

	// blocks until the muxado connnection is closed
	session.Wait()

	// delete client
	//delete(clients, session)
}

func main() {
	fmt.Println(banner)
	var err error
	cfg, err = loadConfig()
	if err != nil {
		logger.Fatal("Error loading config", err)
		return
	}

	typ := "tls"
	if len(os.Args) >= 2 {
		typ = os.Args[1]
	}
	logger.Debug("Using Connection type: ", typ)

	//should/could possibly be lower
	//schedule(updatejson, 1000*time.Millisecond)
	//schedule(func() { log.Println(clientjson) }, 2000*time.Millisecond)

	// go startHTTP(cfg.HTTPPort)
	// logger.Info("Started HTTP server at port ", cfg.HTTPPort)

	// run main tcp server
	listen(typ)
}
