package main

import (
	"os"

	"github.com/amitbet/teleporter/client"
	"github.com/amitbet/teleporter/logger"
)

func main() {
	typ := "tls"
	if len(os.Args) < 2 {
		logger.Error("Invalid arguments.\n cln.exe <address:port> \nexample usage: cln.exe 127.0.0.1:8484")
		os.Exit(0)
	}
	if len(os.Args) >= 3 {
		typ = os.Args[2]
		logger.Info("Using connection type: ", typ)
	}
	serverAddress := os.Args[1]
	sess := client.NewClientConnection(serverAddress, typ)
	client.HandleClientConnection(sess)
}
