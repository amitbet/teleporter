package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/amitbet/teleporter/logger"
	"github.com/inconshreveable/muxado"
	"github.com/pions/dtls/pkg/dtls"
)

var cfg *config

var clientjson []byte

var clients = make(map[muxado.Session]*client)

const banner = `Teleproxy server!`

func createListener(listenerType string) (net.Listener, error) {
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

//Listen create a listener and serve on it
func listen(listenerType string) error {

	controlListener, err := createListener(listenerType)
	if err != nil {
		return err
	}

	socksAddr := ":" + cfg.SocksPort
	var socksListener net.Listener
	//Setup tcp server on a known port

	socksListener, err = net.Listen("tcp", socksAddr)
	if err != nil {
		log.Println("Error starting Socks listener", err)
		return err
	}
	logger.Info("Started new SOCKS listener at port %v\n", socksListener.Addr().String())

	//pass := "b2whr9" // randomString(6)
	defer controlListener.Close()
	for {
		conn, err := controlListener.Accept()
		if err != nil {
			logger.Error("TCP accept failed: %s\n", err)
			continue
		}
		go handle(conn, socksListener, cfg.SocksPass)
	}
}

// handle manages a new physical client connection comming into the control port
func handle(conn net.Conn, socksListener net.Listener, pass string) {
	// Setup server side of muxado
	session := muxado.Server(conn, nil)
	defer session.Close()

	//create new Client
	client := newClient(session, socksListener, cfg.SocksUsername, pass)
	clients[session] = client
	go client.listen()

	//blocks until the muxado connnection is closed
	client.wait()

	//delete client
	delete(clients, session)
}

func main() {
	fmt.Println(banner)
	var err error
	cfg, err = loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	typ := "tls"
	if len(os.Args) >= 2 {
		typ = os.Args[1]
	}
	logger.Debug("Using Connection type: ", typ)

	//should/could possibly be lower or smth
	schedule(updatejson, 1000*time.Millisecond)
	//schedule(func() { log.Println(clientjson) }, 2000*time.Millisecond)

	go startHTTP(cfg.HTTPPort)
	log.Println("Started HTTP server at port " + cfg.HTTPPort)

	//run main tcp server
	listen(typ)
}
