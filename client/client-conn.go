package client

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"os"

	"github.com/amitbet/teleporter/common"
	"github.com/amitbet/teleporter/logger"
	"github.com/armon/go-socks5"
	"github.com/pions/dtls/pkg/dtls"
)

// NewClientConnection creates a new client and opens multiple connections to the given server
func NewClientConnection(serverAddress string, connType string) *MultiMux {

	host, _ := os.Hostname()
	config := common.ClientConfig{
		ClientId:       host,
		NetworkExports: []string{"*"},
		Mapping:        make(map[string]string),
	}
	config.Mapping[""] = ""
	jstr, err := json.Marshal(config)
	if err != nil {
		logger.Error("problem in client config json marshaling: ", err)
	}

	sess := NewMultiMux(true)
	for i := 0; i < 10; i++ {
		conn1 := dialConnection(connType, serverAddress)

		// write the client ID & Configuration to the server
		common.WriteString(conn1, string(jstr))
		sess.AddConnection(conn1)
	}
	return sess
}

// HandleClientConnection runs the accept loop on the client side mux, serving any incoming requests
func HandleClientConnection(sess *MultiMux) {

	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	for {
		sconn, err := sess.Accept()
		if err != nil {
			logger.Error("Can't accept, connection is dead", err)
			break
		}
		logger.Debug("mux connection accepted")
		go server.ServeConn(sconn)
	}
}

// dialConnection opens a single connection to the server
func dialConnection(typ string, serverAddress string) net.Conn {

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	var err error
	var conn net.Conn
	if typ == "tls" {
		conn, err = tls.Dial("tcp", serverAddress, tlsconfig)
		if err != nil {
			logger.Error("Cannot connect to target: ", err)
			os.Exit(0)
		}
	} else if typ == "dtls" {
		logger.Debug("running on dtls")
		addr, err := net.ResolveUDPAddr("udp", serverAddress)
		if err != nil {
			logger.Error("Cannot resolve address: ", serverAddress, err)
			os.Exit(0)
		}

		//addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4444}

		// Generate a certificate and private key to secure the connection
		certificate, privateKey, err := dtls.GenerateSelfSigned()
		if err != nil {
			panic(err)
		}

		// Prepare the configuration of the DTLS connection
		config := &dtls.Config{Certificate: certificate, PrivateKey: privateKey}

		// Connect to a DTLS server
		conn, err = dtls.Dial("udp", addr, config)
	}
	return conn
}
