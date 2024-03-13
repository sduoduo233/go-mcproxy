package core

import (
	"bufio"
	"log"
	"mcproxy/config"
	"net"
	"strings"
)

var cfg config.Config

func Start(c config.Config) {
	cfg = c

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		log.Fatalln("listen error:", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("accept error:", err)
			return
		}
		go handler(conn)
	}
}

func handler(conn net.Conn) {
	defer conn.Close()
	defer log.Println("connection end:", conn.RemoteAddr().String())
	log.Println("new connection:", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	defer reader.Reset(nil)

	pkt, err := ReadPacket(reader)
	if err != nil {
		log.Println("read packet error:", pkt)
		return
	}

	// packet handshake
	var protocol VarInt
	var address String
	var port UShort
	var nextState VarInt
	_, err = pkt.Scan(&protocol, &address, &port, &nextState)
	if err != nil {
		log.Println("scan handshake error:", err)
		return
	}

	log.Printf("address=%s:%d, protocol=%d, state=%d", address, port, protocol, nextState)

	switch nextState {
	case 1: // status

		err := handlePing(reader, conn, int(protocol))
		if err != nil {
			log.Println("handle ping error:", err)
		}

	case 2: // login

		if protocol < VERSION_1_8_9 {
			err := sendDisconnect(conn, "unsupported client version")
			if err != nil {
				log.Println("disconnect error:", err)
			}
			return
		}

		// disconnect if server is full
		if onlineCount.Load() >= int32(cfg.MaxPlayer) {
			err := sendDisconnect(conn, "The server is full")
			if err != nil {
				log.Println("disconnect error:", err)
			}
			return
		}

		err := handleForward(reader, conn, strings.HasSuffix(string(address), "\x00FML\x00"), int(protocol))
		if err != nil {
			log.Println("handle forward error:", err)
		}

	}
}
