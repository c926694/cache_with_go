package main

import (
	"cache"
	"flag"
	"log"
)

func main() {
	addr := flag.String("addr", "0.0.0.0:3000", "server listen address")
	capacity := flag.Int("capacity", 1024*1024*64, "cache capacity in bytes")

	flag.Parse()

	server := cache.NewServer(*capacity, *addr)

	log.Printf("cache server listening on %s, capacity=%d bytes", *addr, *capacity)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
