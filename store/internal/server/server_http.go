package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Cloud-RAMP/kv-store/store/internal/store"
)

type logEntry struct {
	method  string
	key     string
	success bool
	time    time.Time
}

var logChan chan logEntry = make(chan logEntry, 500)

func Start(address string) error {
	http.HandleFunc("/", handler)
	fmt.Println("Server listening on", address)

	// dedicated logging chan
	go func() {
		for log := range logChan {
			fmt.Printf("Method:%s Key:%s Success:%t Time:%s\n", log.method, log.key, log.success, log.time.Format("2006-01-02 15:04:05"))
		}
	}()

	return http.ListenAndServe(address, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:] // Remove leading "/"
	switch r.Method {
	case "GET":
		handleGet(w, key)
	case "POST":
		handlePut(w, r, key)
	case "DELETE":
		handleDel(w, key)
	default:
		http.Error(w, "Invalid method. Send GET POST or DELETE", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, key string) {
	val, err := store.Get(key)
	if err != nil {
		http.Error(w, "Key not found", 404)
		logRequest(false, "GET", key)
		return
	}

	w.Write([]byte(val))
	logRequest(true, "GET", key)
}

func handlePut(w http.ResponseWriter, r *http.Request, key string) {
	var body struct {
		Value string `json:"value"` //this syntax makes for easy JSON parsing
	}

	if json.NewDecoder(r.Body).Decode(&body) != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	store.Put(key, body.Value)
	logRequest(true, "PUT", key)
	w.WriteHeader(200)
}

func handleDel(w http.ResponseWriter, key string) {
	store.Del(key)
	logRequest(true, "DELETE", key)
	w.WriteHeader(200)
}

func logRequest(success bool, method string, key string) {
	logChan <- logEntry{
		method:  method,
		key:     key,
		success: success,
		time:    time.Now(),
	}
}
