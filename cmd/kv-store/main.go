package main

import "github.com/Cloud-RAMP/kv-store/internal/server"

func main() {
	server.Start(":3000")
}
