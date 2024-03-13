package core

import (
	"encoding/json"
	"fmt"
	"io"
	"sync/atomic"
)

const VERSION_1_8_9 = 47
const VERSION_1_18_2 = 758

var onlineCount atomic.Int32

// write disconnect packet
func sendDisconnect(w io.Writer, reason string) error {
	type chat struct {
		Text string `json:"text"`
	}

	bytes, err := json.Marshal(&chat{
		Text: reason,
	})
	if err != nil {
		return err
	}

	pkt, err := Pack(String(string(bytes)))
	if err != nil {
		return fmt.Errorf("pack disconnect: %w", err)
	}

	return WritePacket(0x00, pkt, w)
}

type statusVersion struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type statusPlayers struct {
	Max    int        `json:"max"`
	Online int        `json:"online"`
	Sample []struct{} `json:"sample"`
}

type statusResponse struct {
	Version     statusVersion `json:"version"`
	Players     statusPlayers `json:"players"`
	Description string        `json:"description"`
	Favicon     string        `json:"favicon"`
}

// write ping response packet
func sendResponse(w io.Writer, protocol int) error {
	resp, err := json.Marshal(statusResponse{
		Version: statusVersion{
			Name:     "gomcproxy",
			Protocol: protocol,
		},
		Players: statusPlayers{
			Max:    cfg.MaxPlayer,
			Online: int(onlineCount.Load()),
		},
		Description: cfg.Description,
		Favicon:     "",
	})

	if err != nil {
		return fmt.Errorf("response marshal: %w", err)
	}

	pkt, err := Pack(String(resp))
	if err != nil {
		return fmt.Errorf("response pack: %w", err)
	}

	return WritePacket(0x00, pkt, w)
}
