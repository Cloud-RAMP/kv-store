package store

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"
)

const NUM_LOCKS = 11
const SAVE_INTERVAL = 5 * time.Second
const SAVE_FILE_PATH = "internal/store/save.gob"

type lockTable []sync.Mutex
type storeType struct {
	store     []map[string]string
	lockTable lockTable
}

var store storeType

// runs on project initialization
func init() {
	storeTable := make([]map[string]string, NUM_LOCKS)
	for i := range storeTable {
		storeTable[i] = make(map[string]string)
	}

	store = storeType{
		store:     storeTable,
		lockTable: make([]sync.Mutex, NUM_LOCKS),
	}

	// start background goroutine to periodically save the store
	ticker := time.NewTicker(SAVE_INTERVAL)
	go func() {
		for range ticker.C {
			if err := store.save(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: periodic save failed: %v\n", err)
			}
		}
	}()
}

// functions to help with locking (distribute load over NUM_LOCKS locks for quicker access times)
func basicHash(key string) int {
	hash := 0
	for i := 0; i < len(key); i++ {
		hash = (hash*31 + int(key[i])) % NUM_LOCKS
	}

	return hash
}

func (t *lockTable) lock(key string) {
	keyHash := basicHash(key)
	(*t)[keyHash].Lock()
}

func (t *lockTable) unlock(key string) {
	keyHash := basicHash(key)
	(*t)[keyHash].Unlock()
}

// internal store object get/put/del functions

func (s *storeType) get(key string) (string, error) {
	hash := basicHash(key)

	s.lockTable.lock(key)
	defer s.lockTable.unlock(key)

	val, ok := s.store[hash][key]
	if !ok {
		return "", fmt.Errorf("key not found in table")
	}

	return val, nil
}

func (s *storeType) put(key string, val string) {
	hash := basicHash(key)

	s.lockTable.lock(key)
	defer s.lockTable.unlock(key)

	s.store[hash][key] = val
}

func (s *storeType) del(key string) {
	hash := basicHash(key)

	s.lockTable.lock(key)
	defer s.lockTable.unlock(key)

	delete(s.store[hash], key)
}

// save serializes the store to disk using gob encoding.
// All shard locks are held during serialization to ensure a consistent snapshot.
func (s *storeType) save() error {
	// lock all shards in index order to get a consistent snapshot
	for i := range s.lockTable {
		s.lockTable[i].Lock()
	}
	defer func() {
		for i := range s.lockTable {
			s.lockTable[i].Unlock()
		}
	}()

	f, err := os.Create(SAVE_FILE_PATH)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Printf("Saving store to disk at %s\n", SAVE_FILE_PATH)
	return gob.NewEncoder(f).Encode(s.store)
}

// External public facing get/put/del functions
// Denoted as public because they start with a capital letter

func Get(key string) (string, error) {
	return store.get(key)
}

func Put(key string, val string) {
	store.put(key, val)
}

func Del(key string) {
	store.del(key)
}
