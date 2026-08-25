package main

import (
	"log"
)

func main() {
	cfg := LoadConfig()
	store := NewPassStore()
	log.Printf("satellite pass planning service listening on :%s", cfg.Port)
	if err := serveAddress(":"+cfg.Port, NewHandler(store)); err != nil {
		log.Fatal(err)
	}
}
