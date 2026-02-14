package fsm

import "testing"

func TestMachine_RegisterAndLookup(t *testing.T) {
	m := MustNewMachine("transfer-intent")

	start := State("PENDING")
	ev := Event("SUBMIT")
	next := State("SUBMITTED")

	err := m.Register(Transition{From: start, On: ev, To: next, Name: "submit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := m.Lookup(start, ev)
	if !ok {
		t.Fatalf("expected transition to exist")
	}
	if got.To != next {
		t.Fatalf("expected To=%s got=%s", next, got.To)
	}
}

func TestMachine_DuplicateRegisterRejected(t *testing.T) {
	m := MustNewMachine("transfer-intent")

	t1 := Transition{From: State("A"), On: Event("X"), To: State("B")}
	t2 := Transition{From: State("A"), On: Event("X"), To: State("C")}

	if err := m.Register(t1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Register(t2); err != ErrDuplicateTransition {
		t.Fatalf("expected ErrDuplicateTransition, got %v", err)
	}
}
