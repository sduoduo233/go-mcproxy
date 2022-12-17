package main

import (
	"encoding/json"
	"fmt"

	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int           `json:"max"`
		Online int           `json:"online"`
		Sample []interface{} `json:"sample"`
	} `json:"players"`
	Description string `json:"description"`
	Favicon     string `json:"favicon,omitempty"`
}

type PacketHandshake struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       int32
}

type PacketLoginStart struct {
	Name string
}

type PacketStatusResponse struct {
	Response string
}

func WriteStatusResponse(conn *net.Conn, p PacketStatusResponse) error {
	return conn.WritePacket(packet.Marshal(
		0x00,
		packet.String(p.Response),
	))
}

func WriteHandshake(conn *net.Conn, p PacketHandshake) error {
	pkt := packet.Marshal(
		0x00,
		packet.VarInt(p.ProtocolVersion),
		packet.String(p.ServerAddress),
		packet.Short(p.ServerPort),
		packet.VarInt(p.NextState),
	)
	return conn.WritePacket(pkt)
}

func WriteStatusRequest(conn *net.Conn) error {
	return conn.WritePacket(packet.Marshal(
		0x00,
	))
}

func WriteLoginStart(conn *net.Conn, p PacketLoginStart) error {
	pkt := packet.Marshal(
		0x00,
		packet.String(p.Name),
	)
	return conn.WritePacket(pkt)
}

func WriteDisconnect(conn *net.Conn, reason string) error {
	return conn.WritePacket(packet.Marshal(
		0x00,
		packet.String(fmt.Sprintf("{\"text\":\"%s\"}", reason)),
	))
}

func ReadStatusResponse(conn *net.Conn) (*StatusResponse, error) {
	var p packet.Packet

	err := conn.ReadPacket(&p)
	if err != nil {
		return nil, err
	}

	if p.ID != 0x00 {
		return nil, fmt.Errorf("except packet StatusResponse, got %d", p.ID)
	}

	var payload packet.String
	err = p.Scan(&payload)
	if err != nil {
		return nil, err
	}

	var resp StatusResponse
	err = json.Unmarshal([]byte(payload), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func ReadLoginStart(conn *net.Conn) (*PacketLoginStart, error) {
	var p packet.Packet

	err := conn.ReadPacket(&p)
	if err != nil {
		return nil, err
	}

	if p.ID != 0x00 {
		return nil, fmt.Errorf("except packet LoginStart, got %d", p.ID)
	}

	var name packet.String

	err = p.Scan(&name)
	if err != nil {
		return nil, err
	}

	return &PacketLoginStart{
		Name: string(name),
	}, nil
}

func ReadHandshake(conn *net.Conn) (*PacketHandshake, error) {
	var p packet.Packet

	err := conn.ReadPacket(&p)
	if err != nil {
		return nil, err
	}

	if p.ID != 0x00 {
		return nil, fmt.Errorf("except packet Handshake, got %d", p.ID)
	}

	var (
		protocolVersion packet.VarInt
		serverAddr      packet.String
		serverPort      packet.Short
		nextState       packet.VarInt
	)

	err = p.Scan(
		&protocolVersion,
		&serverAddr,
		&serverPort,
		&nextState,
	)
	if err != nil {
		return nil, err
	}

	return &PacketHandshake{
		ProtocolVersion: int32(protocolVersion),
		ServerAddress:   string(serverAddr),
		ServerPort:      uint16(serverPort),
		NextState:       int32(nextState),
	}, nil
}
