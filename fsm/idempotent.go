package fsm

import (
	"errors"
	"time"
)

// TransitionLogStore abstracts idempotency storage.
//
// First principles:
// - fsmkit must not depend on DB/Redis/etc.
// - caller provides a store implementation.
type TransitionLogStore interface {
	// Get returns a previously stored log for this key.
	// ok=false means not found.
	Get(key string) (log TransitionLog, ok bool, err error)

	// Put stores the log for this key.
	// Must be safe if called multiple times with the same (key, log).
	Put(key string, log TransitionLog) error
}

// IdempotencyKeyFunc derives a stable key.
//
// Good defaults are things like:
// - (machineID + eventID)
// - (machineName + machineID + externalEventID)
// - (machineID + fingerprint of event payload) [if you have one]
type IdempotencyKeyFunc func(machineName, machineID string, from State, on Event, at time.Time, meta map[string]string, input any) (string, error)

// IdempotentMachine wraps a Machine and prevents duplicate application
// when retries occur (distributed systems).
type IdempotentMachine struct {
	M     *Machine
	Store TransitionLogStore
	KeyFn IdempotencyKeyFunc
}

// ErrMissingIdempotencyKey indicates the KeyFn returned empty key.
var ErrMissingIdempotencyKey = errors.New("missing idempotency key")

func NewIdempotentMachine(m *Machine, store TransitionLogStore, keyFn IdempotencyKeyFunc) (*IdempotentMachine, error) {
	if m == nil {
		return nil, ErrNilMachine
	}
	if store == nil {
		return nil, ErrNilStore
	}
	if keyFn == nil {
		return nil, ErrNilKeyFunc
	}
	return &IdempotentMachine{
		M:     m,
		Store: store,
		KeyFn: keyFn,
	}, nil
}

func MustNewIdempotentMachine(m *Machine, store TransitionLogStore, keyFn IdempotencyKeyFunc) *IdempotentMachine {
	im, err := NewIdempotentMachine(m, store, keyFn)
	if err != nil {
		panic(err)
	}
	return im
}

// Apply executes Next() exactly-once per idempotency key.
// If the key was seen, it returns the previously stored log (deterministically).
func (im *IdempotentMachine) Apply(
	machineID string,
	from State,
	on Event,
	at time.Time,
	meta map[string]string,
	input any,
) (State, TransitionLog, error) {

	key, err := im.KeyFn(im.M.name, machineID, from, on, at, meta, input)
	if err != nil {
		return "", TransitionLog{}, err
	}
	if key == "" {
		return "", TransitionLog{}, ErrMissingIdempotencyKey
	}

	// 1) check store first
	if prior, ok, err := im.Store.Get(key); err != nil {
		return "", TransitionLog{}, err
	} else if ok {
		// Important: next state is encoded in the log's To if Allowed
		if prior.Allowed {
			return prior.To, prior, nil
		}
		return "", prior, ErrIllegalTransition
	}

	// 2) compute fresh
	next, log, err := im.M.Next(machineID, from, on, at, meta, input)

	// 3) store log regardless of allowed/blocked (idempotency cares about attempts)
	if putErr := im.Store.Put(key, log); putErr != nil {
		// return storage error but preserve the transition outcome info
		return next, log, putErr
	}

	return next, log, err
}
