package main

import (
	"encoding/json"
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

// forward connection to real server
func forwardConnection(conn mcnet.Conn, handshake PacketHandshake) error {
	remoteConn, err := mcnet.DialMC(*remote)
	if err != nil {
		return err
	}
	defer remoteConn.Close()

	// modify & send handshake packet
	if strings.Contains(handshake.ServerAddress, "\x00FML\x00") {
		handshake.ServerAddress = strings.SplitN(*remote, ":", 2)[0] + "\u0000FML\u0000"
	} else {
		handshake.ServerAddress = strings.SplitN(*remote, ":", 2)[0]
	}
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
