package main

import (
	"bufio"
	"log"
	"os"

	mcnet "github.com/Tnze/go-mc/net"
	"golang.org/x/net/proxy"
)

type Socks5Proxy struct {
	Address  string
	Username string
	Password string
}

var proxies = []Socks5Proxy{}

// load proxies from file
func loadProxies() {
	file, err := os.Open(*socks5)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxies = append(proxies, Socks5Proxy{
			Address: scanner.Text(),
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Loaded %d proxies", len(proxies))
}

// returns a new connection to the real server
func dialMC() (*mcnet.Conn, error) {
	if len(proxies) == 0 {
		// dial without proxy
		log.Printf("connecting to %s directly", *remote)

		remoteConn, err := mcnet.DialMC(*remote)
		if err != nil {
			return nil, err
		}
		return remoteConn, nil
	} else {
		// dial using proxy
		log.Printf("connecting to %s using socks5 proxy %s", *remote, proxies[0].Address)

		remoteConn, err := dialScosk5(*remote, proxies[0].Address, proxies[0].Username, proxies[0].Password)
		proxies = proxies[1:]

		if err != nil {
			return nil, err
		}
		return remoteConn, nil
	}
}

func dialScosk5(addr string, proxyAddr string, username string, password string) (*mcnet.Conn, error) {
	auth := proxy.Auth{
		User:     username,
		Password: password,
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, &auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return mcnet.WrapConn(conn), nil
}
