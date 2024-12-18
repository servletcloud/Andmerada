package main

import (
	"log"

	"github.com/servletcloud/Andmerada/internal/cmd"
)

func main() {
	if err := cmd.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
