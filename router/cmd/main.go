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
		log.Fatalf("invalid KV_NODES: %v", err)
	}

	if err := router.Start(addr, nodes); err != nil {
		log.Fatalf("router failed: %v", err)
	}
}
