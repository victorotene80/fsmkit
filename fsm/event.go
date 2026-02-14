package fsm

import "strings"

// Event is an opaque identifier for a transition trigger.
type Event string

func (e Event) String() string { return string(e) }

// Normalize makes event deterministic for matching.
func (e Event) Normalize() Event {
	return Event(strings.TrimSpace(string(e)))
}

func (e Event) Valid() bool {
	v := strings.TrimSpace(string(e))
	if v == "" || len(v) > 64 {
		return false
	}
	return validIdent(v)
}
