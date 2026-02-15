package fsm

import "errors"

var (
	ErrInvalidState       = errors.New("invalid state")
	ErrInvalidEvent       = errors.New("invalid event")
	ErrInvalidTransition  = errors.New("invalid transition")
	ErrInvalidMachineName = errors.New("invalid machine name")

	ErrNoTransition        = errors.New("no transition for state+event")
	ErrDuplicateTransition = errors.New("duplicate transition for state+event")
	ErrIllegalTransition   = errors.New("illegal transition")

	ErrNilMachine = errors.New("nil machine")
	ErrNilStore   = errors.New("nil store")
	ErrNilKeyFunc = errors.New("nil idempotency key func")
)
