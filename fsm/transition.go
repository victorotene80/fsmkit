package fsm

import (
	"sort"
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

func (l TransitionLog) CanonicalMetaPairs() []string {
	if len(l.Meta) == 0 {
		return nil
	}
	keys := make([]string, 0, len(l.Meta))
	for k := range l.Meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+l.Meta[k])
	}
	return out
}

func (l TransitionLog) CanonicalString() string {
	// deterministic string; no spaces; fixed field order
	// time is RFC3339Nano in UTC
	var b strings.Builder
	b.Grow(128)

	b.WriteString("from=")
	b.WriteString(l.From.String())
	b.WriteString("|on=")
	b.WriteString(l.On.String())
	b.WriteString("|to=")
	b.WriteString(l.To.String())
	b.WriteString("|at=")
	b.WriteString(l.At.UTC().Format(time.RFC3339Nano))
	b.WriteString("|allowed=")
	if l.Allowed {
		b.WriteString("1")
	} else {
		b.WriteString("0")
	}
	b.WriteString("|reason=")
	b.WriteString(string(l.Reason))

	pairs := l.CanonicalMetaPairs()
	for _, p := range pairs {
		b.WriteString("|m=")
		b.WriteString(p)
	}
	return b.String()
}
