package fsm

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type transitionKey struct {
	from State
	on   Event
}

func key(from State, on Event) transitionKey {
	return transitionKey{from: from, on: on}
}

// Machine is a reusable finite state machine definition.
// It stores allowed transitions. Execution is pure (Next returns state+log).
type Machine struct {
	name string

	initial State
	hasInit bool

	transitions map[transitionKey]Transition
}

func NewMachine(name string) (*Machine, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 128 {
		return nil, ErrInvalidMachineName
	}
	return &Machine{
		name:        name,
		transitions: make(map[transitionKey]Transition),
	}, nil
}

func MustNewMachine(name string) *Machine {
	m, err := NewMachine(name)
	if err != nil {
		panic(err)
	}
	return m
}

func (m *Machine) Name() string { return m.name }

func (m *Machine) SetInitial(s State) error {
	s = s.Normalize()
	if !s.Valid() {
		return ErrInvalidState
	}
	m.initial = s
	m.hasInit = true
	return nil
}

func (m *Machine) Initial() (State, bool) {
	return m.initial, m.hasInit
}

// Register adds a transition rule. (From, On) must be unique.
func (m *Machine) Register(t Transition) error {
	t = t.Normalize()
	if !t.Valid() {
		return ErrInvalidTransition
	}

	k := key(t.From, t.On) // <-- MUST use normalized fields
	if _, exists := m.transitions[k]; exists {
		return ErrDuplicateTransition
	}

	m.transitions[k] = t // <-- store normalized t
	return nil
}

func (m *Machine) MustRegister(t Transition) {
	if err := m.Register(t); err != nil {
		panic(err)
	}
}

func (m *Machine) Lookup(from State, on Event) (Transition, bool) {
	from = from.Normalize()
	on = on.Normalize()
	t, ok := m.transitions[key(from, on)]
	return t, ok
}

// Transitions returns a deterministic snapshot.
func (m *Machine) Transitions() []Transition {
	out := make([]Transition, 0, len(m.transitions))
	for _, t := range m.transitions {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.On != b.On {
			return a.On < b.On
		}
		return a.To < b.To
	})
	return out
}

// Next evaluates the transition (including guard) and returns:
// - next state (if allowed)
// - TransitionLog (always)
// - error (ErrNoTransition / ErrIllegalTransition / input errors)
func (m *Machine) Next(
	machineID string,
	from State,
	on Event,
	at time.Time,
	meta map[string]string,
	input any,
) (State, TransitionLog, error) {

	// Normalize first (so " pending " works deterministically)
	from = from.Normalize()
	on = on.Normalize()

	at = at.UTC()
	if meta == nil {
		meta = map[string]string{}
	}

	// validate inputs
	if !from.Valid() {
		log := TransitionLog{
			MachineID: machineID,
			From:      from,
			On:        on,
			To:        "",
			At:        at,
			Meta:      meta,
			Allowed:   false,
			Reason:    ReasonInvalidInput,
		}
		return "", log, ErrInvalidState
	}
	if !on.Valid() {
		log := TransitionLog{
			MachineID: machineID,
			From:      from,
			On:        on,
			To:        "",
			At:        at,
			Meta:      meta,
			Allowed:   false,
			Reason:    ReasonInvalidInput,
		}
		return "", log, ErrInvalidEvent
	}

	t, ok := m.Lookup(from, on)
	if !ok {
		log := TransitionLog{
			MachineID: machineID,
			From:      from,
			On:        on,
			To:        "",
			At:        at,
			Meta:      meta,
			Allowed:   false,
			Reason:    ReasonNoTransition,
		}
		return "", log, ErrNoTransition
	}

	// run guard if present
	if t.Guard != nil {
		ctx := GuardContext{
			MachineName: m.name,
			MachineID:   machineID,
			From:        t.From,
			On:          t.On,
			To:          t.To,
			At:          at,
			Meta:        meta,
			Input:       input,
		}

		if err := t.Guard.Check(ctx); err != nil {
			if _, blocked := isBlocked(err); blocked {
				log := TransitionLog{
					MachineID: machineID,
					From:      t.From,
					On:        t.On,
					To:        t.To,
					At:        at,
					Meta:      meta,
					Allowed:   false,
					Reason:    ReasonGuardBlocked,
				}
				return "", log, ErrIllegalTransition
			}

			// unexpected guard failure
			log := TransitionLog{
				MachineID: machineID,
				From:      t.From,
				On:        t.On,
				To:        t.To,
				At:        at,
				Meta:      meta,
				Allowed:   false,
				Reason:    ReasonInternalError,
			}
			return "", log, errors.Join(ErrIllegalTransition, err)
		}
	}

	// allowed
	log := TransitionLog{
		MachineID: machineID,
		From:      t.From,
		On:        t.On,
		To:        t.To,
		At:        at,
		Meta:      meta,
		Allowed:   true,
		Reason:    ReasonOK,
	}
	return t.To, log, nil
}
