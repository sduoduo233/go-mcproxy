package main

import (
	"encoding/base64"
	"log"
	"os"
)

func base64Encode(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("failed to load favicon ", path, err)
		return ""
	}

	s := "data:image/png;base64,"
	s += base64.StdEncoding.EncodeToString(bytes)
	return s
}
