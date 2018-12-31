package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/armon/go-socks5"
	"github.com/pions/dtls/pkg/dtls"
)

func main() {
	typ := "tls"
	if len(os.Args) < 2 {
		fmt.Println("Invalid arguments.\n cln.exe <address:port> \nexample usage: cln.exe 127.0.0.1:8484")
		os.Exit(0)
	}
	if len(os.Args) >= 3 {
		typ = os.Args[2]
		fmt.Println("found type: ", typ)
	}

	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	serverAddress := os.Args[1]

	// conn := dialConnection(typ, serverAddress)
	// sess := NewMultiMuxClient(conn)

	//---- without multimux
	// conn := dialConnection(typ, serverAddress)
	// sess := muxado.Client(conn, nil)

	sess := NewMultiMuxClient1()
	for i := 0; i < 10; i++ {
		conn1 := dialConnection(typ, serverAddress)
		sess.AddConnection(conn1)
	}

	for {
		sconn, err := sess.Accept()
		if err != nil {
			log.Println("Can't accept, connection is dead", err)
			break
		}
		fmt.Println("mux connection accepted")
		go server.ServeConn(sconn)
	}
	// Simple way to keep program running until CTRL-C is pressed.
	//<-make(chan struct{})
}
func dialConnection(typ string, serverAddress string) net.Conn {

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	var err error
	var conn net.Conn
	if typ == "tls" {
		conn, err = tls.Dial("tcp", serverAddress, tlsconfig)
		if err != nil {
			log.Println("Cannot connect to target: ", err)
			os.Exit(0)
		}
	} else if typ == "dtls" {
		fmt.Println("running on dtls")
		addr, err := net.ResolveUDPAddr("udp", serverAddress)
		if err != nil {
			log.Println("Cannot resolve address: ", serverAddress, err)
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
