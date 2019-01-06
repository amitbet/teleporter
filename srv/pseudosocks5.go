package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

//code mostly ripped from github.com/armon/go-socks5
const (
	socks5Version = uint8(5)

	NoAuth          = uint8(0)
	noAcceptable    = uint8(255)
	UserPassAuth    = uint8(2)
	userAuthVersion = uint8(1)
	authSuccess     = uint8(0)
	authFailure     = uint8(1)
)

// readMethods is used to read the number of methods
// and proceeding auth methods
func readMethods(r io.Reader) ([]byte, error) {
	header := []byte{0}
	if _, err := r.Read(header); err != nil {
		return nil, err
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(r, methods, numMethods)
	return methods, err
}

func getauthdata(reader io.Reader, writer io.Writer) (string, string, error) {

	// Get the version and username length
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
		return "", "", err
	}

	// Ensure we are compatible
	if header[0] != userAuthVersion {
		return "", "", fmt.Errorf("Unsupported auth version: %v", header[0])
	}

	// Get the user name
	userLen := int(header[1])
	user := make([]byte, userLen)
	if _, err := io.ReadAtLeast(reader, user, userLen); err != nil {
		return "", "", err
	}

	// Get the password length
	if _, err := reader.Read(header[:1]); err != nil {
		return "", "", err
	}

	// Get the password
	passLen := int(header[0])
	pass := make([]byte, passLen)
	if _, err := io.ReadAtLeast(reader, pass, passLen); err != nil {
		return "", "", err
	}

	//log.Println("Auth: " + string(user) + " " + string(pass))
	return string(user), string(pass), nil
}

// auth is used to handle connection authentication
func auth(conn net.Conn, username string, password string) error {
	bufConn := bufio.NewReader(conn)

	// Read the version byte
	version := []byte{0}
	if _, err := bufConn.Read(version); err != nil {
		err := fmt.Errorf("[AUTH] socks: Failed to get version byte: %v", err)
		return err
	}

	// Ensure we are compatible
	if version[0] != socks5Version {
		err := fmt.Errorf("[AUTH] Unsupported SOCKS version: %v", version)
		return err
	}

	// Get the methods
	methods, err := readMethods(bufConn)
	if err != nil {
		return fmt.Errorf("[AUTH] Failed to get auth methods: %v", err)
	}

	// Select a usable method (only auth for us here)
	for _, method := range methods {
		if method == UserPassAuth {
			// Tell the client to use user/pass auth
			if _, err := conn.Write([]byte{socks5Version, UserPassAuth}); err != nil {
				return fmt.Errorf("[AUTH] Can't write method reply: %v", err)
			}

			//read username/password
			u, p, err := getauthdata(bufConn, conn)
			//logger.Debugf("user: %s, pass: %s",u, p)
			if (err != nil) || (u != username || p != password) {
				conn.Write([]byte{userAuthVersion, authFailure})
				return fmt.Errorf("[AUTH] Username/password auth failed: %v", err)
			}

			//correct auth
			conn.Write([]byte{userAuthVersion, authSuccess})
			return nil
		}
	}

	// No usable method found
	conn.Write([]byte{socks5Version, noAcceptable})
	return fmt.Errorf("[AUTH] No acceptable auth method")
}
