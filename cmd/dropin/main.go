package main

import (
	"flag"
	"log"

	"github.com/kamaln7/dropin-chatops"
	"github.com/kamaln7/dropin-chatops/config"
)

var (
	configPath string
)

func main() {
	flag.StringVar(&configPath, "config", "./config.toml", "path to config file")
	flag.Parse()

	conf, err := config.Read(configPath)
	if err != nil {
		log.Fatalf("error reading config: %v\n", err)
	}

	bot := dropin.New(conf)
	if err := bot.Serve(); err != nil {
		log.Println(err)
	}
}
