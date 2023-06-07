package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Listen      string   `json:"listen"`
	Description string   `json:"description"`
	Remote      string   `json:"remote"`
	Favicon     string   `json:"favicon"`
	MaxPlayer   int      `json:"max_player"`
	PingMode    string   `json:"ping_mode"` // fake, real
	FakePing    int      `json:"fake_ping"`
	RewirteHost string   `json:"rewrite_host"`
	RewirtePort int      `json:"rewrite_port"`
	Auth        string   `json:"auth"` // none, whitelist, blacklist
	Whitelist   []string `json:"whitelist"`
	Blacklist   []string `json:"blacklist"`
}

func ParseConfig(path string) *Config {
	config := Config{
		Listen:      "0.0.0.0:25565",
		Description: "hello\nworld",
		Remote:      "mc.hypixel.net:25565",
		Favicon:     "",
		MaxPlayer:   20,
		PingMode:    "fake",
		FakePing:    0,
		RewirteHost: "mc.hypixel.net",
		RewirtePort: 25565,
		Auth:        "none",
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config %s: %s", path, err)
		return nil
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatalf("invalid json: %s", err)
		return nil
	}

	if config.PingMode != "fake" && config.PingMode != "real" {
		log.Fatalf("invalid ping_mode in config: %s", config.PingMode)
		return nil
	}

	if config.Auth != "none" && config.Auth != "blacklist" && config.Auth != "whitelist" {
		log.Fatalf("invalid auth in config: %s", config.PingMode)
		return nil
	}

	return &config
}
