package agent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/amitbet/go-socks5"
	"github.com/amitbet/teleporter/logger"
	proxy_dialer "github.com/mwitkow/go-http-dialer"
	//"github.com/pions/dtls"
)

type IMux interface {
	Open() (net.Conn, error)
	Accept() (net.Conn, error)
	AddConnection(c io.ReadWriteCloser)
}

// Tether is a generalized network connection for tunneling
// it usually holds a bundle of connections, multiplexed into endless virtual connections
// along with any additional information about the remote part of the call
type Tether struct {
	IMux
	RemoteConfig *ClientConfig
}

// NewTether creates a tether which is a generalized network connection for tunneling
func NewTether(isClient bool) *Tether {
	t := Tether{}
	t.IMux = NewMultiMux(isClient)
	return &t
}

// Router holds all connections for the current snap-node, along with the network configuration & routing logic
// it recieves network connections and routes them to the correct destination
type Router struct {
	socks5server       *socks5.Server
	socks5Credentials  socks5.StaticCredentials
	tethers            map[string]*Tether
	NetworkConfig      *ClientConfig
	mu                 sync.RWMutex
	AuthenticateSocks5 bool
	Proxy              *ProxyInfo
}

func NewRouter() *Router {
	rtr := &Router{}
	rtr.AuthenticateSocks5 = true
	conf := &socks5.Config{}
	s5server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	rtr.socks5server = s5server
	//rtr.IncomingConns = make(chan *server.TunnelTask, 16)
	rtr.tethers = make(map[string]*Tether)

	//load & populate network configuration
	host, _ := os.Hostname()
	config := ClientConfig{
		ClientId: host,
		// NetworkExports: []string{"*"},
		Mapping: make(map[string]string),
	}
	//config.Mapping[""] = ""
	rtr.NetworkConfig = &config

	return rtr
}

func GenerateSocks5Req(task *TaskInfo) *socks5.Request {

	var ipv4Address = uint8(1)
	var fqdnAddress = uint8(3)
	var ipv6Address = uint8(4)

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
			Port:        0,
		},
	}

	intPort, err := strconv.Atoi(task.TargetPort)
	if err != nil {
		logger.Error("GenerateSocks5Req: error in converting port to int:", err)
	}

	req.DestAddr.Port = int(intPort)
	addr := net.ParseIP(task.TargetAddress)
	if addr != nil && len(addr) != 0 {
		req.DestAddr.IP = addr
		if len(req.DestAddr.IP) == 4 {
			req.DestAddr.AddressType = ipv4Address
		} else {
			req.DestAddr.AddressType = ipv6Address
		}
	} else {
		req.DestAddr.AddressType = fqdnAddress
		req.DestAddr.FQDN = task.TargetAddress
	}

	return &req
}

// getTargetTether finds the path for a given request (task) and returns the next tether through which it should be routed
func (rtr *Router) getTargetTether(taskInf *TaskInfo) (*Tether, error) {
	var tID string

	//search our network mapping for any explicit routes
	for wildcardStr, targetID := range rtr.NetworkConfig.Mapping {
		regStr := strings.Replace(wildcardStr, "*", ".*?", -1)
		reg := regexp.MustCompile(regStr)
		if reg.MatchString(taskInf.TargetAddress) {
			tID = targetID
			break
		}
	}

	// for any targets that should be locally executed return nil (local execution)
	if rtr.NetworkConfig.ClientId == tID || // we found our own name in the map
		strings.ToLower(tID) == "local" || // we have an explicit local in the map
		strings.ToLower(tID) == "localhost" ||
		(tID == "" && taskInf.Local) { // we didn't find anything explicit in the map but the client is a local-only socks5 listener
		logger.Debug("Router.route: Executing locally for target: " + taskInf.TargetAddress + ":" + taskInf.TargetPort)
		return nil, nil
	}

	logger.Debug("Router.route: Found route to: " + tID + " for target: " + taskInf.TargetAddress + ":" + taskInf.TargetPort)

	//lookup the tether by its id:
	rtr.mu.RLock()
	teth, ok := rtr.tethers[tID]
	rtr.mu.RUnlock()

	//if not found - there is no route, send back an error..
	if !ok {
		errorStr := "thether not found in router.getTargetTether: " + tID
		logger.Error(errorStr)
		return nil, errors.New(errorStr)
	}
	return teth, nil
}

// route contains the logic which decides where to send the network task once it is acquired
// it then either relays the task to another node, piping the connections together,
// or executes the task in the local network
func (rtr *Router) route(task *TunnelTask) {
	// if task.Header.Type == TaskTypePing {
	// 	task.Conn.Write([]byte("Pong"))
	// 	return
	// }

	teth, err := rtr.getTargetTether(task.Header)
	if err != nil {
		//kill task by not relaying it further
		logger.Error("Router.route Error: no thether - disposing of task")
		return
	}

	if teth == nil {
		// ----- if no relay required, execute locally:
		req := GenerateSocks5Req(task.Header)
		b := &bytes.Buffer{}
		req.WriteTo(b)
		task.PrefixSend(b.Bytes())
		//b := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		//task.Read(b)
		//logger.Fatal(b)
		rtr.taskExec(task)
	} else {
		// ----- relay the task to the next node:
		logger.Info("chosen route:", teth.RemoteConfig.ClientId)

		//add the task info to the stream for the other side to route.
		task.PrefixTaskInfo()
		rtr.taskRelay(task, teth)
	}
}

// taskRelay will relay the task to the network node dscribed by the target parameter
func (rtr *Router) taskRelay(task *TunnelTask, targ *Tether) error {
	muxConn, err := targ.Open()
	if err != nil {
		logger.Error("Error establishing session", err)
		return err
	}

	defer muxConn.Close()
	defer task.Conn.Close()

	//send all prebuffered content down the line
	muxConn.Write(task.ReadPresend())

	//From now on proxy everything
	errCh := make(chan error, 2)
	go proxy(task.Conn, muxConn, errCh)
	go proxy(muxConn, task.Conn, errCh)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			logger.Error("Error in io.copy: ", e)
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

type closeWriter interface {
	CloseWrite() error
}

// proxy is used to stream data from src to destination, and sends errors
// down a dedicated channel
func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	q, err := io.Copy(dst, src)
	logger.Infof("transferred %d bytes", q)
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
	errCh <- err
}

// taskExec will run the task with the local server, performing the request inside the current network
// it can have several modes of operation (socks5, vpn, htmlproxy), currently only socks5 is implemented.
func (rtr *Router) taskExec(task *TunnelTask) {
	rtr.executeAsSocks5(task)
}

// Connect creates a new bundle of physical connections to the server (AKA: thether)
func (rtr *Router) Connect(connConf *TetherConfig, numConnsPerTether int) error {
	proxy := connConf.Proxy
	serverAddress := connConf.TargetHost + ":" + strconv.Itoa(connConf.TargetPort)
	if numConnsPerTether <= 0 {
		numConnsPerTether = 10
	}

	teth, err := rtr.createMultiConn(serverAddress, connConf, numConnsPerTether, proxy)
	if err != nil {
		logger.Error("Connect: problem while connecting the tether to server:", serverAddress, err)
		return err
	}

	if strings.TrimSpace(teth.RemoteConfig.ClientId) == "" {
		logger.Error("Connect: bad clientID while connecting tether to server:", serverAddress)
		return fmt.Errorf("Connect: bad clientID while connecting tether to server: %s", serverAddress)
	}

	rtr.mu.Lock()
	rtr.tethers[teth.RemoteConfig.ClientId] = teth
	rtr.mu.Unlock()
	go rtr.handleIncomingConnections(teth)
	return err
}

// Serve creates a listener of given type and runs it on the given port
func (rtr *Router) Serve(serverConf ListenerConfig) error {
	port := strconv.Itoa(serverConf.Port)
	switch serverConf.Type {
	case "socks5": // opens a socks 5 proxy port for browsers / native clients
		// an entry point for incoming traffic
		socksAddr := ":" + port
		if serverConf.LocalOnly {
			socksAddr = "localhost" + socksAddr
		}
		creds := serverConf.AuthorizedClients

		rtr.AuthenticateSocks5 = serverConf.UseAuthentication
		rtr.socks5Credentials = creds
		socks5Listener, err := rtr.createSocks5Listener(socksAddr)
		if err != nil {
			logger.Error("problem with listening to port: ", port, err)
			return err
		}
		go rtr.handleSocksListener(socks5Listener)
	case "relayTcp": // opens a multi-mux tcp port, executes locally or realys messages to other connections
		// tcp is a solid default to start from
		listenAddr := ":" + port
		controlListener, err := createTlsControlListener(listenAddr)
		if err != nil {
			logger.Error("problem with listening to port: ", port, err)
			return err
		}
		go rtr.handleControlListener(controlListener, &serverConf)
	case "relayUdp":
		// udp is good for performance
		// listenAddr := ":" + port
		// controlListener1, err := createDtlsControlListener(listenAddr)
		// if err != nil {
		// 	logger.Error("problem with listening to port: ", port, err)
		// 	return err
		// }
		// go rtr.handleControlListener(controlListener1, &serverConf)
	case "relayWebSockets":
		// ws is good for passing firewalls
		return errors.New("Not implemented")
	default:
		return errors.New("Unknown server type: " + serverConf.Type)
	}
	return nil
}

func (rtr *Router) handleControlListener(controlListener net.Listener, serverConf *ListenerConfig) {
	defer controlListener.Close()
	for {
		conn, err := controlListener.Accept()
		if err != nil {
			logger.Error("TCP accept failed: %s\n", err)
			continue
		}
		go rtr.handlePhysicalClientConn(conn, serverConf)
	}
}

func readNetConfig(conn net.Conn) (*ClientConfig, error) {
	clientConfigStr, err := ReadString(conn)
	if err != nil {
		logger.Error("Client connect, failed while reading client header: %s\n", err)
		return nil, err
	}

	logger.Debug("client connected, read client config string: ", clientConfigStr)
	cconfig := ClientConfig{}
	err = json.Unmarshal([]byte(clientConfigStr), &cconfig)
	if err != nil {
		logger.Error("Client connect, error unmarshaling clientConfig: %s\n", err)
		return nil, err
	}
	return &cconfig, nil
}

func writeNetConfig(conn net.Conn, config *ClientConfig) error {
	jstr, err := json.Marshal(config)
	if err != nil {
		logger.Error("writeNetConfig: problem in netConfig json marshaling: ", err)
		return err
	}
	err = WriteString(conn, string(jstr))
	if err != nil {
		logger.Error("writeNetConfig: Problem in sending network config: ", err)
		return err
	}
	return nil
}

// handlePhysicalClientConn manages a new physical (non-mux) client connection comming into the control port
// it reads the client configuration, answers with our node's config, and adds the connection to the correct multi-mux conn pool
func (rtr *Router) handlePhysicalClientConn(conn net.Conn, serverConf *ListenerConfig) {

	err := writeNetConfig(conn, rtr.NetworkConfig)
	if err != nil {
		logger.Error("handlePhysicalClientConn: error writing netConfig", err)
		return
	}

	// read client configuration from conn
	cconfig, err := readNetConfig(conn)
	if err != nil {
		logger.Error("handlePhysicalClientConn: error reading netConfig from client", err)
		return
	}

	cid := cconfig.ClientId
	logger.Info("Client connected, id: ", cid)

	if serverConf.UseAuthentication {
		if serverConf.AuthorizedClients[cid] != cconfig.Secret {
			logger.Warn("Authentication error, bad password for clientId: " + cid)
			conn.Close()
			return
		}
	}
	// Setup server side of muxado
	// session := muxado.Server(conn, nil)
	// defer session.Close()

	var teth *Tether
	var ok bool
	rtr.mu.RLock()
	teth, ok = rtr.tethers[cid]
	rtr.mu.RUnlock()

	//TODO: go over goroutines, and inspect creation and deletion

	// if this is a first connection to some node in the netowrk, create a new tether to represent it
	if !ok {
		//create new Client
		teth = NewTether(false)
		teth.RemoteConfig = cconfig
		rtr.mu.Lock()
		rtr.tethers[cid] = teth
		rtr.mu.Unlock()

		//TODO: cleanup and close all connections when listener is destroyed, think of dead conns and reconnect
		//TODO: check that incoming mux conns are closed and that go routines handling them end as expected
		//TODO: add some keepalive mechanism
		teth.AddConnection(conn)
		go rtr.handleIncomingConnections(teth)
	} else {
		// a known node, add an additional connection to an existing tether
		teth.AddConnection(conn)
		teth.RemoteConfig = cconfig
	}

}

// createTlsControlListener creates a listener of type: tcpRelay (with encryption = tls)
func createTlsControlListener(listenAddress string) (net.Listener, error) {
	var controlListener net.Listener
	var err error

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return nil, err
	}
	tlsconfig := &tls.Config{Certificates: []tls.Certificate{cer}}

	controlListener, err = tls.Listen("tcp", listenAddress, tlsconfig)

	if err != nil {
		return nil, err
	}

	logger.Debug("createControlListener: Started server at " + listenAddress)
	return controlListener, nil
}

// createDtlsControlListener creates a listener of type: udpRelay (with encryption = dtls)
// func createDtlsControlListener(listenAddress string) (net.Listener, error) {
// 	var controlListener net.Listener
// 	var err error

// 	logger.Debug("createControlListener: running on dtls")
// 	addr, err := net.ResolveUDPAddr("udp", listenAddress)
// 	if err != nil {
// 		return nil, err
// 	}

// 	certificate, privateKey, err := dtls.GenerateSelfSigned()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Prepare the configuration of the DTLS connection
// 	config := &dtls.Config{Certificate: certificate, PrivateKey: privateKey}

// 	// Connect to a DTLS server
// 	controlListener, err = dtls.Listen("udp", addr, config)

// 	if err != nil {
// 		return nil, err
// 	}
// 	logger.Debug("createControlListener: Started server at " + listenAddress)
// 	return controlListener, nil
// }

func (rtr *Router) executeAsSocks5(muxConn *TunnelTask) {
	// read request from connection:
	request, err := socks5.NewRequest(muxConn)
	if err != nil {
		logger.Error("Router.executeAsSocks5: Error: ", err)
		return
	}

	defer muxConn.Close()

	// Process the client request
	if err := rtr.socks5server.HandleRequest(request, muxConn); err != nil {
		logger.Error("Failed to handle request", err)
		return
	}

	return
	//start socks5 connection handler

	//read the auth answer
	//reply := []byte{0, 0}
	//io.ReadAtLeast(muxConn.Conn, reply, 2)
}

// runPingLoop periodically pings the other side
// and listen for the reply, which should be recieved within a certain time period
// func runPingLoop(intervalSecs, timeoutSecs int, conn net.Conn) error {
// 	err := writeTaskInfo(conn, &TaskInfo{
// 		Type:          TaskTypePing,
// 		TargetPort:    "",
// 		TargetAddress: "",
// 		Local:         false,
// 	})
// 	if err != nil {
// 		logger.Error("error while sending ping header: ", err)
// 		return err
// 	}
// 	pong := []byte{0, 0, 0, 0}
// 	conn.SetDeadline(time.Now().Add(time.Duration(timeoutSecs) * time.Second))
// 	_, err = conn.Read(pong)
// 	conn.SetDeadline(nil)
// 	return nil
// }

// createMultiConn opens multiple connections to the given server
func (rtr *Router) createMultiConn(serverAddress string, tConf *TetherConfig, connCountInBundle int, proxyInfo *ProxyInfo) (*Tether, error) {
	myConf := *rtr.NetworkConfig
	myConf.Secret = tConf.ClientPassword

	jstr, err := json.Marshal(myConf)
	if err != nil {
		logger.Error("createMultiConn: problem in network config json marshaling: ", err)
	}

	th := NewTether(true)
	for i := 0; i < connCountInBundle; i++ {
		conn1 := dialConnection(tConf.ConnectionType, serverAddress, proxyInfo)

		// read ID & config from the client
		cconfig, err := readNetConfig(conn1)
		if err != nil {
			logger.Error("createMultiConn: problem in reading client's network config: ", err)
			return nil, err
		}
		th.RemoteConfig = cconfig

		// write the client ID & Configuration to the server
		err = WriteString(conn1, string(jstr))
		if err != nil {
			logger.Error("createMultiConn: problem in sending server's network config: ", err)
			return nil, err
		}

		th.AddConnection(conn1)
	}
	return th, nil
}

// HandleClientConnection runs the accept loop on the client side multi-mux (tether),
// serving any incoming requests
func (rtr *Router) handleIncomingConnections(sess *Tether) {

	for {
		sconn, err := sess.Accept()
		if err != nil {
			logger.Error("Can't accept, connection is dead", err)
			break
		}
		logger.Debug("mux connection accepted")
		task, err := ReadTunnelTask(sconn)
		if err != nil {
			logger.Error("failed to read task from connection", err)
			break
		}

		go rtr.route(task)
	}
}

// dialConnection opens a single connection to the server
func dialConnection(typ string, serverAddress string, proxy *ProxyInfo) net.Conn {

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	var err error
	var conn net.Conn
	if typ == "tls" {

		if proxy != nil {
			u, err := url.Parse(proxy.Address)
			if err != nil {
				logger.Error("failed parsing httpProxyListener into URL", err)
			}

			proxyDialer := proxy_dialer.New(u, proxy_dialer.WithTls(tlsconfig))
			if proxy.User != "" || proxy.Pass != "" {
				prxAuth := proxy_dialer.WithProxyAuth(proxy_dialer.AuthBasic(proxy.User, proxy.Pass))
				prxAuth(proxyDialer)
			}

			rawConn, err := proxyDialer.Dial("tcp", serverAddress)
			if err != nil {
				logger.Error("Cannot connect to target: ", err)
				os.Exit(0)
			}
			conn = tls.Client(rawConn, tlsconfig)

			if err != nil {
				logger.Error("Cannot estabilsh tls: ", err)
				os.Exit(0)
			}
		} else {
			conn, err = tls.Dial("tcp", serverAddress, tlsconfig)
		}

		if err != nil {
			logger.Error("Cannot connect to target: ", err)
			os.Exit(0)
		}
	} else if typ == "dtls" {
		// logger.Debug("running on dtls")
		// addr, err := net.ResolveUDPAddr("udp", serverAddress)
		// if err != nil {
		// 	logger.Error("Cannot resolve address: ", serverAddress, err)
		// 	os.Exit(0)
		// }

		// // Generate a certificate and private key to secure the connection
		// certificate, privateKey, err := dtls.GenerateSelfSigned()
		// if err != nil {
		// 	panic(err)
		// }

		// // Prepare the configuration of the DTLS connection
		// config := &dtls.Config{Certificate: certificate, PrivateKey: privateKey}

		// // Connect to a DTLS server
		// conn, err = dtls.Dial("udp", addr, config)
	}
	return conn
}

func (rtr *Router) createSocks5Listener(socksAddr string) (net.Listener, error) {

	socksListener, err := net.Listen("tcp", socksAddr)
	if err != nil {
		logger.Error("Error starting Socks listener", err)
		return nil, err
	}
	logger.Infof("Started new SOCKS listener at port %v\n", socksListener.Addr().String())
	return socksListener, nil
}

func (rtr *Router) handleSocks5Connection(conn net.Conn) {
	var cator socks5.Authenticator
	// 5s to get the socks establishing over with
	///conn.SetDeadline(time.Now().Add(5 * time.Second))
	if rtr.AuthenticateSocks5 {
		// connect & authenticate
		cator = socks5.UserPassAuthenticator{
			Credentials: rtr.socks5Credentials,
		}
	} else {
		cator = socks5.NoAuthAuthenticator{}
	}
	req, err := socks5.PerformHandshake(conn, []socks5.Authenticator{cator})

	if err != nil {
		logger.Error("Error in socks5 handshake: ", err)
		return
	}

	address := req.DestAddr.Address()

	// u, err := url.Parse(address)
	// if err != nil {
	// 	panic(err)
	// }
	destHost, destPort, _ := net.SplitHostPort(address)

	// adding the request to this task, so it can be handled on the other side -----
	buff := bytes.Buffer{}
	req.WriteTo(&buff)

	task := NewTunnelTask(
		conn,
		&TaskInfo{
			Type:          TaskTypeSocks,
			TargetPort:    destPort,
			TargetAddress: destHost,
			Local:         true,
		})

	rtr.route(task)
}

// handles an incomming socks connection
func (rtr *Router) handleSocksListener(listener net.Listener) {
	for {
		// Accept a TCP connection
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Closed tcp server: ", err)
			continue
		}
		go rtr.handleSocks5Connection(conn)
	}
}
