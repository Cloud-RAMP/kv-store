package server

// this was a work in progress. it would be really fast but hard to parse headers, etc.
// could work on this later

// import (
// 	"bufio"
// 	"fmt"
// 	"net"

// 	"github.com/Cloud-RAMP/kv-store/internal/store"
// )

// // need to make proper HTTP responses later
// // could do the whole server in HTTP, but slower

// func Start(address string) error {
// 	listener, err := net.Listen("tcp", address)
// 	if err != nil {
// 		return err
// 	}
// 	defer listener.Close()
// 	fmt.Println("Server listening on", address)

// 	// infinite loop, handle connections
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			fmt.Println("Accept error:", err)
// 			continue
// 		}

// 		// spawn a new user-level thread process to handle the connection
// 		go handleConnection(conn)
// 	}
// }

// func handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	reader := bufio.NewReader(conn)
// 	method, err := reader.ReadString(' ')
// 	if err != nil {
// 		fmt.Printf("handle error later here")
// 		return
// 	}

// 	// cut off trailing space
// 	method = method[:len(method)-1]
// 	fmt.Println("New request:", method)

// 	// operate according to request method
// 	switch method {
// 	case "GET":
// 		handleGet(conn, reader)
// 		return
// 	case "POST":
// 		handlePut(conn, reader)
// 		return
// 	case "DEL":
// 		handleDel(conn, reader)
// 		return
// 	}

// 	conn.Write([]byte("some proper http response here for invalid method"))
// }

// func handleGet(conn net.Conn, reader *bufio.Reader) {
// 	key, err := reader.ReadString(' ')
// 	if err != nil {
// 		conn.Write([]byte("http 500 error"))
// 		return
// 	}

// 	// this could fail on empty key but it's fast
// 	key = key[:len(key)-1][1:]
// 	fmt.Println("requested key:", key)

// 	val, err := store.Get(key)
// 	if err != nil {
// 		conn.Write([]byte("i think 404 error here? not sure"))
// 		return
// 	}

// 	conn.Write([]byte(val))
// }

// func handlePut(conn net.Conn, reader *bufio.Reader) {
// 	key, err := reader.ReadString(' ')
// 	if err != nil {
// 		conn.Write([]byte("http 500 error"))
// 		return
// 	}

// 	// this could fail on empty key but it's fast
// 	key = key[:len(key)-1][1:]
// 	fmt.Println("requested key:", key)
// }

// func handleDel(conn net.Conn, reader *bufio.Reader) {
// 	fmt.Println("handling del")

// }
