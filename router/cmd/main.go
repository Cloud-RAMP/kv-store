package main

import (
	"log"
	"os"

	"github.com/Cloud-RAMP/kv-store/router/internal/router"
)

func main() {
	addr := os.Getenv("ROUTER_ADDR")
	if addr == "" {
		addr = ":3000"
	}

	nodes, err := router.ParseNodes(os.Getenv("KV_NODES"))
	if err != nil {
		nodes = []string{
			"http://localhost:3001", "http://localhost:3002", "http://localhost:3003",
		}
	}

	if err := router.Start(addr, nodes); err != nil {
		log.Fatalf("router failed: %v", err)
	}
}
