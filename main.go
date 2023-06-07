package main

import (
	"flag"
	"log"
	"mcproxy/config"
	"mcproxy/core"
)

const version = "2.0.0"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("gomcproxy (version %s)", version)

	configPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	config := config.ParseConfig(*configPath)
	core.Start(*config)
}
