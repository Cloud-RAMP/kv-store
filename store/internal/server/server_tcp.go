package server

// import (
// 	"bufio"
// 	"fmt"
// 	"net"
// 	"strings"

// 	"github.com/Cloud-RAMP/kv-store/store/internal/store"
// )

// func Start(address string) error {
// 	listener, err := net.Listen("tcp", address)
// 	if err != nil {
// 		return err
// 	}
// 	defer listener.Close()
// 	fmt.Println("Server listening on", address)

// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			fmt.Println("Accept error:", err)
// 			continue
// 		}

// 		go handleConnection(conn)
// 	}
// }

// func handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	reader := bufio.NewReader(conn)
// 	line, err := reader.ReadString('\n')
// 	if err != nil {
// 		conn.Write([]byte("ERROR: failed to read request\n"))
// 		return
// 	}

// 	// Parse the line: METHOD key [value]
// 	parts := strings.Fields(strings.TrimSpace(line))
// 	if len(parts) < 2 {
// 		conn.Write([]byte("ERROR: invalid format\n"))
// 		return
// 	}

// 	method := parts[0]
// 	key := parts[1]

// 	switch method {
// 	case "GET":
// 		handleGet(conn, key)
// 	case "POST":
// 		if len(parts) < 3 {
// 			conn.Write([]byte("ERROR: POST requires key and value\n"))
// 			return
// 		}
// 		value := parts[2]
// 		handlePut(conn, key, value)
// 	case "DELETE":
// 		handleDel(conn, key)
// 	default:
// 		conn.Write([]byte("ERROR: invalid method\n"))
// 	}
// }

// func handleGet(conn net.Conn, key string) {
// 	val, err := store.Get(key)
// 	if err != nil {
// 		conn.Write([]byte("ERROR: key not found\n"))
// 		return
// 	}
// 	conn.Write([]byte(val + "\n"))
// }

// func handlePut(conn net.Conn, key string, value string) {
// 	store.Put(key, value)
// 	conn.Write([]byte("OK\n"))
// }

// func handleDel(conn net.Conn, key string) {
// 	store.Del(key)
// 	conn.Write([]byte("OK\n"))
// }
