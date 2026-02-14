package fsm

import "strings"

// State is an opaque identifier for a machine state.
type State string

func (s State) String() string { return string(s) }

// Normalize makes state deterministic for matching.
func (s State) Normalize() State {
	return State(strings.TrimSpace(string(s)))
}

func (s State) Valid() bool {
	v := strings.TrimSpace(string(s))
	if v == "" || len(v) > 64 {
		return false
	}
	return validIdent(v)
}
