package core

import (
	"fmt"
	"io"
	"log"
	"sync"
)

func handleForward(reader io.Reader, writer io.Writer, fml bool, protocol int) error {
	onlineCount.Add(1)
	defer onlineCount.Add(-1)

	// read login start
	pkt, err := ReadPacket(reader)
	if err != nil {
		return fmt.Errorf("read pkt login start: %w", err)
	}
	if pkt.ID != 0x00 {
		return fmt.Errorf("expect packet login start, got %d", pkt.ID)
	}

	var username String
	_, err = pkt.Scan(&username)
	if err != nil {
		return fmt.Errorf("scan login start: %w", err)
	}

	log.Printf("user login: %s", username)

	// whitelist / blacklist
	allow, msg, err := allowJoin(string(username))
	if err != nil {
		log.Printf("authentication failed: %s", err)
		return nil
	}

	if !allow {
		log.Printf("user rejected: %s, reason: %s", username, msg)

		err = sendDisconnect(writer, msg)
		if err != nil {
			return fmt.Errorf("write disconnect: %w", err)
		}
		return nil
	}

	// connect to remote
	remote, err := DialMC(cfg.Remote)
	if err != nil {
		return err
	}
	defer remote.Close()

	// handshake packet
	rewriteHost := cfg.RewirteHost
	if fml {
		rewriteHost += "\x00FML\x00"
	}

	pktHandshake, err := Pack(
		VarInt(protocol),
		String(rewriteHost),
		UShort(cfg.RewirtePort),
		VarInt(2), // next state login
	)
	if err != nil {
		return err
	}
	err = WritePacket(0x00, pktHandshake, remote)
	if err != nil {
		return fmt.Errorf("write handshake: %w", err)
	}

	// write login start
	pktLoginStart, err := Pack(
		String(username),
	)
	if err != nil {
		return fmt.Errorf("pack login start: %w", err)
	}
	err = WritePacket(0x00, pktLoginStart, remote)
	if err != nil {
		return fmt.Errorf("write login start: %w", err)
	}

	// start forward
	log.Println("start forward:", username)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(writer, remote)
		wg.Done()
	}()
	go func() {
		io.Copy(remote, reader)
		wg.Done()
	}()

	wg.Wait()
	return nil
}
