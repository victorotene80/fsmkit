package fsm

import "errors"

var (
	ErrInvalidState       = errors.New("invalid state")
	ErrInvalidEvent       = errors.New("invalid event")
	ErrInvalidTransition  = errors.New("invalid transition")
	ErrInvalidMachineName = errors.New("invalid machine name")

	ErrNoTransition        = errors.New("no transition for state+event")
	ErrDuplicateTransition = errors.New("duplicate transition for state+event")
)
