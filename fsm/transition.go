package fsm

import (
	"strings"
	"time"
)

type Transition struct {
	From State
	On   Event
	To   State

	Name  string
	Guard Guard // optional
}

func (t Transition) Normalize() Transition {
	t.From = t.From.Normalize()
	t.On = t.On.Normalize()
	t.To = t.To.Normalize()
	t.Name = strings.TrimSpace(t.Name)
	return t
}

func (t Transition) Valid() bool {
	return t.From.Valid() && t.On.Valid() && t.To.Valid()
}

type ReasonCode string

const (
	ReasonOK            ReasonCode = "ok"
	ReasonInvalidInput  ReasonCode = "invalid_input"
	ReasonNoTransition  ReasonCode = "no_transition"
	ReasonGuardBlocked  ReasonCode = "guard_blocked"
	ReasonInternalError ReasonCode = "internal_error"
)

type TransitionLog struct {
	MachineID string
	From      State
	On        Event
	To        State

	At time.Time

	Meta map[string]string

	Allowed bool
	Reason  ReasonCode
}
