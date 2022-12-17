package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mcnet "github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
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
	listen      = flag.String("listen", "127.0.0.1:25565", "local listening address")
	remote      = flag.String("remote", "mc.hypixel.net:25565", "remote forward address")
	help        = flag.Bool("help", false, "print help message")
	fakePing    = flag.Bool("fakeping", false, "fake ping")
	description = flag.String("description", "", "server description")
	favicon     = flag.String("favicon", "favicon.png", "server icon")
	max         = flag.Int("max", 20, "max player")
)

// forward connection to real server
func forwardConnection(conn mcnet.Conn, handshake PacketHandshake) error {
	remoteConn, err := mcnet.DialMC(*remote)
	if err != nil {
		return err
	}

	// modify & send handshake packet
	handshake.ServerAddress = strings.SplitN(*remote, ":", 2)[0]
	port, err := strconv.Atoi(strings.SplitN(*remote, ":", 2)[1])
	if err != nil {
		log.Fatal("invalid port: ", *remote, err)
	}
	handshake.ServerPort = uint16(port)
	WriteHandshake(remoteConn, handshake)

	// read username
	loginStart, err := ReadLoginStart(&conn)
	if err != nil {
		return nil
	}
	log.Println("login:", loginStart.Name)

	// check allowJoin and kick player
	allow, reason := allowJoin(loginStart.Name)
	if !allow {
		return WriteDisconnect(&conn, reason)
	}
	WriteLoginStart(remoteConn, *loginStart)

	// forward connection
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(remoteConn, conn)
		wg.Done()
	}()
	go func() {
		io.Copy(conn, remoteConn)
		wg.Done()
	}()
	wg.Wait()

	return nil
}

// handle ping request
func handlePing(conn mcnet.Conn, handshake PacketHandshake) error {
	log.Println("ping", conn.Socket.RemoteAddr().String())

	// connect to the real server
	remoteConn, err := mcnet.DialMC(*remote)
	if err != nil {
		return err
	}
	defer remoteConn.Close()

	// write handshake to real server
	err = WriteHandshake(remoteConn, PacketHandshake{
		ProtocolVersion: 47,
		ServerAddress:   strings.SplitN(*remote, ":", 2)[0],
		ServerPort:      25565,
		NextState:       NEXTSTATE_STATUS,
	})
	if err != nil {
		return err
	}

	for {
		var p packet.Packet
		err := conn.ReadPacket(&p)
		if err != nil {
			return err
		}

		switch p.ID {
		case 0x00: // status request
			var resp *StatusResponse

			err := WriteStatusRequest(remoteConn)
			if err != nil {
				return err
			}
			resp, err = ReadStatusResponse(remoteConn)
			if err != nil {
				return err
			}

			resp.Version.Name = "go-mcproxy"
			resp.Version.Protocol = 47
			resp.Players.Max = *max
			resp.Players.Online = int(atomic.LoadInt32(&online))
			if *description != "" {
				resp.Description = *description
			}
			if *favicon != "" {
				resp.Favicon = faviconBase64
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				return nil
			}

			err = WriteStatusResponse(&conn, PacketStatusResponse{
				Response: string(bytes),
			})
			if err != nil {
				return err
			}

		case 0x01: // ping
			var payload packet.Long
			err := p.Scan(&payload)
			if err != nil {
				return err
			}

			if !(*fakePing) {
				err = remoteConn.WritePacket(packet.Marshal(
					0x01,
					packet.Long(time.Now().UnixMilli()),
				))
				if err != nil {
					return err
				}
				var pkt packet.Packet
				err = remoteConn.ReadPacket(&pkt)
				if err != nil {
					return err
				}
			}

			err = conn.WritePacket(packet.Marshal(
				0x01,
				packet.Long(payload)),
			)
			if err != nil {
				return err
			}
		}
	}

}

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
