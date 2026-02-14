package fsm

import (
	"errors"
	"time"
)

// GuardContext is deterministic input to a guard.
// No time.Now; caller supplies At.
type GuardContext struct {
	MachineName string
	MachineID   string

	From State
	On   Event
	To   State

	At   time.Time
	Meta map[string]string

	// Arbitrary domain input (opaque to fsmkit).
	Input any
}

// Guard decides if a transition is allowed.
// nil => allow
// Blocked(...) => deny deterministically
// other error => internal guard failure
type Guard interface {
	Check(ctx GuardContext) error
}

// GuardFunc adapter.
type GuardFunc func(ctx GuardContext) error

func (g GuardFunc) Check(ctx GuardContext) error { return g(ctx) }

// blockedError marks a deterministic denial.
type blockedError struct {
	reason string
}

func (e blockedError) Error() string {
	if e.reason == "" {
		return "guard blocked"
	}
	return "guard blocked: " + e.reason
}

// Blocked is how guards deterministically deny a transition.
func Blocked(reason string) error {
	return blockedError{reason: reason}
}

func isBlocked(err error) (blockedError, bool) {
	var b blockedError
	if errors.As(err, &b) {
		return b, true
	}
	return blockedError{}, false
}
