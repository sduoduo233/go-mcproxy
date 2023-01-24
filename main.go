package main

import (
	"flag"
	"log"
	"sync/atomic"

	mcnet "github.com/Tnze/go-mc/net"
)

const (
	NEXTSTATE_STATUS = 1
	NEXTSTATE_LOGIN  = 2
)

var (
	online        = int32(0)
	faviconBase64 = ""
)

var (
	listen         = flag.String("listen", "127.0.0.1:25565", "local listening address")
	remote         = flag.String("remote", "mc.hypixel.net:25565", "remote forward address")
	help           = flag.Bool("help", false, "print help message")
	fakePing       = flag.Bool("fakeping", false, "fake ping")
	description    = flag.String("description", "", "server description")
	favicon        = flag.String("favicon", "favicon.png", "server icon")
	max            = flag.Int("max", 20, "max player")
	socks5         = flag.String("socks5", "socks5.txt", "socks5 proxy")
	socks5Username = flag.String("socks5user", "", "socks5 username")
	socks5Password = flag.String("socks5pass", "", "socks5 password")
)

func handleConnection(conn mcnet.Conn) error {
	log.Println("new connection", conn.Socket.RemoteAddr().String())

	atomic.AddInt32(&online, 1)
	defer atomic.AddInt32(&online, -1)
	defer conn.Close()
	defer log.Println("connection closed", conn.Socket.RemoteAddr().String())

	handshake, err := ReadHandshake(&conn)
	if err != nil {
		return err
	}

	if handshake.NextState == NEXTSTATE_LOGIN {
		err = forwardConnection(conn, *handshake)
		return err
	}
	if handshake.NextState == NEXTSTATE_STATUS {
		err = handlePing(conn, *handshake)
		return err
	}

	return nil
}

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	// load favicon
	if *favicon != "" {
		faviconBase64 = base64Encode(*favicon)
	}

	log.Println("gomcproxy")

	// load proxies
	if *socks5 != "" {
		loadProxies()
	}

	server, err := mcnet.ListenMC(*listen)
	if err != nil {
		log.Fatal("listen error: ", err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal("accept error: ", err)
		}
		go func() {
			err := handleConnection(conn)
			if err != nil {
				log.Println("handle connection error:", err)
			}
		}()
	}
}
