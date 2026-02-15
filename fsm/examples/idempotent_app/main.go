package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/victorotene80/fsmkit/fsm"
)

type MemoryStore struct {
	mu sync.RWMutex
	m  map[string]fsm.TransitionLog
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{m: make(map[string]fsm.TransitionLog)}
}

func (s *MemoryStore) Get(key string) (fsm.TransitionLog, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok, nil
}

func (s *MemoryStore) Put(key string, log fsm.TransitionLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = log
	return nil
}

func main() {

	m := fsm.MustNewMachine("transfer-intent")

	m.MustRegister(fsm.Transition{
		From: fsm.State("PENDING"),
		On:   fsm.Event("SUBMIT"),
		To:   fsm.State("SUBMITTED"),
		Name: "submit",
	})

	store := NewMemoryStore()

	keyFn := func(machineName, machineID string,
		from fsm.State,
		on fsm.Event,
		at time.Time,
		meta map[string]string,
		input any,
	) (string, error) {

		return machineID + ":event:123", nil
	}

	im := fsm.MustNewIdempotentMachine(m, store, keyFn)

	now := time.Now().UTC()

	next, log, err := im.Apply(
		"tx-1",
		"PENDING",
		"SUBMIT",
		now,
		map[string]string{"source": "api"},
		nil,
	)

	fmt.Println("FIRST:")
	fmt.Println("Next:", next)
	fmt.Println("Allowed:", log.Allowed)
	fmt.Println("Error:", err)
	fmt.Println()

	// Retry call (same idempotency key)
	next2, log2, err2 := im.Apply(
		"tx-1",
		"PENDING",
		"SUBMIT",
		now.Add(time.Minute),
		map[string]string{"source": "retry"},
		nil,
	)

	fmt.Println("RETRY:")
	fmt.Println("Next:", next2)
	fmt.Println("Allowed:", log2.Allowed)
	fmt.Println("Error:", err2)
	fmt.Println()

	fmt.Println("Deterministic log:")
	fmt.Println(log2.CanonicalString())
}
