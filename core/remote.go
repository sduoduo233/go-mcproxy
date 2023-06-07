package core

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func Resolve(address string) (string, error) {
	// SRV
	if !strings.Contains(address, ":") {
		_, addrs, err := net.LookupSRV("minecraft", "tcp", address)

		if err != nil || len(addrs) == 0 {
			// use default port if SRV failed
			return net.JoinHostPort(address, "25565"), nil
		}

		return net.JoinHostPort(addrs[0].Target, strconv.Itoa(int(addrs[0].Port))), nil
	}

	return address, nil
}

func DialMC(a string) (net.Conn, error) {
	addr, err := Resolve(a)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
