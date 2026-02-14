package fsm

import (
	"sort"
	"strings"
)

type transitionKey struct {
	from State
	on   Event
}

func key(from State, on Event) transitionKey {
	return transitionKey{from: from, on: on}
}

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

func (m *Machine) Register(t Transition) error {
	t = t.Normalize()
	if !t.Valid() {
		return ErrInvalidTransition
	}

	k := key(t.From, t.On)
	if _, exists := m.transitions[k]; exists {
		return ErrDuplicateTransition
	}

	m.transitions[k] = t
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

func (m *Machine) Transitions() []Transition {
	out := make([]Transition, 0, len(m.transitions))
	for _, t := range m.transitions {
		out = append(out, t)
	}

	// Deterministic ordering for tests/docs/debugging
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
