package main

import (
	"log"

	"whr-im/server/internal/config"
	"whr-im/server/internal/router"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	r := router.New(cfg)
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
