package main

import (
	"log"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/config"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/application"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	application.Run(cfg)
}
