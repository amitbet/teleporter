package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/amitbet/teleporter/common"
)

const (
	ipv4Address = uint8(1)
	fqdnAddress = uint8(3)
	ipv6Address = uint8(4)
)

// AddrSpec is used to return the target AddrSpec
// which may be specified as IPv4, IPv6, or a FQDN
type AddrSpec struct {
	AddressType byte
	FQDN        string
	IP          net.IP
	Port        int
}

func (a *AddrSpec) String() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// Address returns a string suitable to dial; prefer returning IP-based
// address, fallback to FQDN
func (a AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

// A Request represents request received by a server
type Request struct {
	// Protocol version
	Version uint8
	// Requested command
	Command uint8
	// reserved header byte
	Reserved uint8
	// AuthContext provided during negotiation
	//AuthContext *AuthContext
	// AddrSpec of the the network that sent the request
	RemoteAddr *AddrSpec
	// AddrSpec of the desired destination
	DestAddr *AddrSpec
	// AddrSpec of the actual destination (might be affected by rewrite)
	realDestAddr *AddrSpec
	bufConn      io.Reader
}

type conn interface {
	Write([]byte) (int, error)
	RemoteAddr() net.Addr
}

// NewRequest creates a new Request from the tcp connection
func NewRequest(bufConn io.Reader) (*Request, error) {
	// Read the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(bufConn, header, 3); err != nil {
		return nil, fmt.Errorf("Failed to get command version: %v", err)
	}

	// Ensure we are compatible
	if header[0] != socks5Version {
		return nil, fmt.Errorf("Unsupported command version: %v", header[0])
	}

	// Read in the destination address
	dest, err := readAddrSpec(bufConn)
	if err != nil {
		return nil, err
	}

	request := &Request{
		Version:  socks5Version,
		Command:  header[1],
		Reserved: header[2],
		DestAddr: dest,
		bufConn:  bufConn,
	}

	return request, nil
}

// WriteTo writes the request to the given writer
func (r *Request) WriteTo(w io.Writer) error {
	header := []byte{r.Version, r.Command, r.Reserved}
	_, err := w.Write(header)
	if err != nil {
		return err
	}
	return writeAddrSpec(w, r.DestAddr)
}

func writeAddrSpec(w io.Writer, addr *AddrSpec) error {
	// write address type
	_, err := w.Write([]byte{addr.AddressType})
	if err != nil {
		return err
	}
	// write address
	switch addr.AddressType {
	case ipv4Address:
		_, err = w.Write(addr.IP)
	case ipv6Address:
		_, err = w.Write(addr.IP)
	case fqdnAddress:
		err = common.WriteShortString(w, addr.FQDN)
	}
	if err != nil {
		return err
	}

	// write port
	arr := []byte{0, 0}
	binary.BigEndian.PutUint16(arr, uint16(addr.Port))

	// arr[0] = byte(addr.Port & (0xff))
	// b := addr.Port >> 2
	// arr[1] = byte(b & (0xff))
	_, err = w.Write(arr)
	if err != nil {
		return err
	}

	// no error
	return nil
}

// readAddrSpec is used to read AddrSpec.
// Expects an address type byte, follwed by the address and port
func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	d := &AddrSpec{}

	// Get the address type
	addrType := []byte{0}

	if _, err := r.Read(addrType); err != nil {
		return nil, err
	}
	//logger.Debug("type:", addrType)

	d.AddressType = addrType[0]

	// Handle on a per type basis
	switch d.AddressType {
	case ipv4Address:
		addr := make([]byte, 4)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case ipv6Address:
		addr := make([]byte, 16)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case fqdnAddress:
		if _, err := r.Read(addrType); err != nil {
			return nil, err
		}
		//logger.Debug("fdqn len:", addrType)
		addrLen := int(addrType[0])
		fqdn := make([]byte, addrLen)
		if _, err := io.ReadAtLeast(r, fqdn, addrLen); err != nil {
			return nil, err
		}
		//logger.Debug("fdqn:", fqdn)
		d.FQDN = string(fqdn)

	default:
		return nil, fmt.Errorf("Unrecognized address type")
	}

	// Read the port
	port := []byte{0, 0}
	if _, err := io.ReadAtLeast(r, port, 2); err != nil {
		return nil, err
	}
	//logger.Debug("port:", port)
	d.Port = (int(port[0]) << 8) | int(port[1])

	return d, nil
}
