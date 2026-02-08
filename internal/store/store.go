package store

import (
	"fmt"
	"sync"
)

const NUM_LOCKS = 11

type lockTable []sync.Mutex
type storeType struct {
	store     map[string]string
	lockTable lockTable
}

var store storeType

// runs on project initialization
func init() {
	store = storeType{
		store:     make(map[string]string),
		lockTable: make([]sync.Mutex, NUM_LOCKS),
	}
}

// functions to help with locking (distribute load over NUM_LOCKS locks for quicker access times)
func basicHash(key string) int {
	out := 1
	for i := range len(key) {
		out *= int(key[i])
	}
	return out % NUM_LOCKS
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
	s.lockTable.lock(key)
	defer s.lockTable.unlock(key)

	val, ok := s.store[key]
	if !ok {
		return "", fmt.Errorf("key not found in table")
	}

	return val, nil
}

func (s *storeType) put(key string, val string) {
	s.lockTable.lock(key)
	defer s.lockTable.unlock(key)

	s.store[key] = val
}

func (s *storeType) del(key string) {
	s.lockTable.lock(key)
	defer s.lockTable.unlock(key)

	delete(s.store, key)
}

// External public facing get/put/del functions

func Get(key string) (string, error) {
	return store.get(key)
}

func Put(key string, val string) {
	store.put(key, val)
}

func Del(key string) {
	store.del(key)
}
