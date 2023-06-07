package core

import (
	"fmt"
	"io"
	"math"
	"time"
)

func handlePing(reader io.Reader, writer io.Writer) error {

	// request
	pkt, err := ReadPacket(reader)
	if err != nil {
		return err
	}
	if pkt.ID != 0x00 {
		return fmt.Errorf("expect packet Request, got %d", pkt.ID)
	}

	// response
	err = sendResponse(writer)
	if err != nil {
		return err
	}

	// fake ping
	if cfg.PingMode == "fake" {
		time.Sleep(time.Millisecond * time.Duration(cfg.FakePing))

		// pong
		pktBytes, err := Pack(Long(math.MaxInt64))
		if err != nil {
			return fmt.Errorf("pack pong: %w", err)
		}
		err = WritePacket(0x01, pktBytes, writer)
		return err
	}

	// ping
	pkt, err = ReadPacket(reader)
	if err != nil {
		return err
	}
	if pkt.ID != 0x01 {
		return fmt.Errorf("expect packet Ping, got %d", pkt.ID)
	}
	var payload Long
	_, err = pkt.Scan(&payload)
	if err != nil {
		return fmt.Errorf("scan ping: %w", err)
	}

	// pong
	pktBytes, err := Pack(payload)
	if err != nil {
		return fmt.Errorf("pack pong: %w", err)
	}
	err = WritePacket(0x01, pktBytes, writer)
	return err
}
