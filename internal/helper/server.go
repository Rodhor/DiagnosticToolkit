package helper

import (
	"fmt"
	"net"
	"strconv"
)

func DefaultAddr() string {
	return "127.0.0.1:8080"
}

func ParseAddr(arg string) (string, error) {
	// Try to convert just the port
	if port, err := strconv.Atoi(arg); err == nil {
		if port < 1 || port > 65535 {
			return "", fmt.Errorf("port %d is out of range (must be between 1 and 65535", port)
		}
		return net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), nil
	}

	// If not just the port, parse entire string
	// as host and port
	host, portStr, err := net.SplitHostPort(arg)
	if err != nil {
		return "", fmt.Errorf("invalid address %q (expected a port number or HOST:PORT", arg)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", fmt.Errorf("invalid port in %q", arg)
	}

	if port < 1 || port > 65535 {
		return "", fmt.Errorf("port %d is out of range (must be between 1 and 65535)", port)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		return "", fmt.Errorf("only localhost addresses are allowed")
	}
	return net.JoinHostPort(host, portStr), nil
}

func IsAddrAvailable(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("address %s unavaible: %w", addr, err)
	}
	_ = ln.Close()
	return nil
}
