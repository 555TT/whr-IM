package main

import (
	"log"

	"whr-im/server/internal/router"
)

func main() {
	r := router.New()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
