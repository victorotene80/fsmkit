package fsm

import "strings"

type State string

func (s State) String() string { return string(s) }

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
