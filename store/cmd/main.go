package main

import (
	"flag"
	"fmt"

	"github.com/Cloud-RAMP/kv-store/store/internal/server"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "3000", "Port to listen on")
	flag.Parse()

	addr := ":" + port
	fmt.Printf("Starting store on %s\n", addr)
	server.Start(addr)
}
