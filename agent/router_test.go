package agent

import (
	"net"
	"testing"
	"time"

	"github.com/amitbet/go-socks5"
)

const (
	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

func TestRouter(t *testing.T) {
	rtr1 := NewRouter()

	relayPass := GenerateRandomString(32)
	conf1 := ListenerConfig{
		Port:              10101,
		Type:              "relayTcp",
		LocalOnly:         false,
		UseAuthentication: true,
		AuthorizedClients: map[string]string{
			"relayUser": relayPass,
		},
	}

	rtr1.Serve(conf1)
	rtr1.NetworkConfig.ClientId = rtr1.NetworkConfig.ClientId + "1"

	rtr2 := NewRouter()
	rtr2.AuthenticateSocks5 = false
	conf2relay := ListenerConfig{
		Port:              10201,
		Type:              "relayTcp",
		LocalOnly:         false,
		UseAuthentication: true,
		AuthorizedClients: map[string]string{
			rtr1.NetworkConfig.ClientId: relayPass,
		},
	}
	conf2socks := ListenerConfig{
		Port:              10202,
		Type:              "socks5",
		LocalOnly:         true,
		UseAuthentication: false,
	}
	rtr2.Serve(conf2relay)
	rtr2.Serve(conf2socks)
	rtr2.NetworkConfig.ClientId = rtr2.NetworkConfig.ClientId + "2"

	// create a tether between router1 & router2
	rtr2.Connect(
		&TetherConfig{
			TargetPort:     10101,
			TargetHost:     "localhost",
			ConnectionType: "tls",
			ConnectionName: "stam1",
			ClientPassword: relayPass,
		},
		10,
	)

	//buff := bytes.Buffer{}

	req := socks5.Request{
		Version: 5,
		// Requested command
		Command: 1,
		// reserved header byte
		Reserved: 0,
		// AuthContext provided during negotiation
		AuthContext: &socks5.AuthContext{
			Method:  0,
			Payload: map[string]string{},
		},
		// AddrSpec of the the network that sent the request
		RemoteAddr: &socks5.AddrSpec{}, // will not be sent anyway
		// AddrSpec of the desired destination
		DestAddr: &socks5.AddrSpec{
			AddressType: ipv4Address, //IPv4 type
			FQDN:        "",          // unimportant
			IP:          net.IP{127, 0, 0, 1},
			Port:        8080,
		},
	}
	_ = req

	//connect to socks5 port
	// // conn, err := net.Dial("tcp", "localhost:10202")
	// // if err != nil {
	// // 	logger.Error("Cannot connect to target: ", err)
	// // 	os.Exit(0)
	// // }

	// // // authenticate with the server
	// // conn.Write([]byte{5, 1, socks5.NoAuth})
	// // reply := []byte{0, 0}
	// // io.ReadAtLeast(conn, reply, 2)

	// // req.WriteTo(conn)

	time.Sleep(time.Second * 400)
}

// func connectToControl() {
// 	tlsconfig := &tls.Config{
// 		InsecureSkipVerify: true,
// 	}

// 	//connect to socks5 port
// 	conn, err := tls.Dial("tcp", "localhost:10202", tlsconfig)
// 	if err != nil {
// 		logger.Error("Cannot connect to target: ", err)
// 		os.Exit(0)
// 	}

// 	// create a network config
// 	host, _ := os.Hostname()
// 	config := common.ClientConfig{
// 		ClientId:       host,
// 		NetworkExports: []string{"*"},
// 		Mapping:        make(map[string]string),
// 	}
// 	config.Mapping[""] = "*"

// 	rtrConf, err := readNetConfig(conn)
// 	err = writeNetConfig(conn, &config)
// 	_ = err
// 	_ = rtrConf
// 	//sess := muxado.Client(conn,nil)
// 	//sess.Open

// }
