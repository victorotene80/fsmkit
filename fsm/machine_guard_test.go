package fsm

import (
	"testing"
	"time"
)

func TestMachine_Next_GuardAllows(t *testing.T) {
	m := MustNewMachine("transfer-intent")
	m.MustRegister(Transition{
		From:  State("PENDING"),
		On:    Event("SUBMIT"),
		To:    State("SUBMITTED"),
		Name:  "submit",
		Guard: GuardFunc(func(ctx GuardContext) error { return nil }),
	})

	next, log, err := m.Next("tx_1", State("PENDING"), Event("SUBMIT"), time.Now(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != State("SUBMITTED") {
		t.Fatalf("expected SUBMITTED got %s", next)
	}
	if !log.Allowed || log.Reason != ReasonOK {
		t.Fatalf("expected allowed log; allowed=%v reason=%s", log.Allowed, log.Reason)
	}
}

func TestMachine_Next_GuardBlocks(t *testing.T) {
	m := MustNewMachine("transfer-intent")
	m.MustRegister(Transition{
		From: State("PENDING"),
		On:   Event("SUBMIT"),
		To:   State("SUBMITTED"),
		Guard: GuardFunc(func(ctx GuardContext) error {
			return Blocked("signatures_invalid")
		}),
	})

	next, log, err := m.Next("tx_1", State("PENDING"), Event("SUBMIT"), time.Now(), nil, map[string]any{"x": 1})
	if err != ErrIllegalTransition {
		t.Fatalf("expected ErrIllegalTransition got %v", err)
	}
	if next != "" {
		t.Fatalf("expected empty next state")
	}
	if log.Allowed || log.Reason != ReasonGuardBlocked {
		t.Fatalf("expected blocked log; allowed=%v reason=%s", log.Allowed, log.Reason)
	}
}

func TestMachine_Next_NoTransition(t *testing.T) {
	m := MustNewMachine("transfer-intent")

	_, log, err := m.Next("tx_1", State("PENDING"), Event("SUBMIT"), time.Now(), nil, nil)
	if err != ErrNoTransition {
		t.Fatalf("expected ErrNoTransition got %v", err)
	}
	if log.Reason != ReasonNoTransition {
		t.Fatalf("expected ReasonNoTransition got %s", log.Reason)
	}
}

func TestMachine_Next_NormalizesTrim(t *testing.T) {
	m := MustNewMachine("transfer-intent")
	m.MustRegister(Transition{
		From: State("PENDING"),
		On:   Event("SUBMIT"),
		To:   State("SUBMITTED"),
	})

	next, _, err := m.Next("tx_1", State(" PENDING "), Event(" SUBMIT "), time.Now(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != State("SUBMITTED") {
		t.Fatalf("expected SUBMITTED got %s", next)
	}
}

func TestStateEvent_InvalidCharactersRejected(t *testing.T) {
	if State("BAD STATE").Valid() {
		t.Fatalf("expected invalid state (space not allowed)")
	}
	if Event("BAD@EVENT").Valid() {
		t.Fatalf("expected invalid event (@ not allowed)")
	}
}

func TestMachine_Next_NormalizesInputs(t *testing.T) {
	m := MustNewMachine("transfer-intent")
	m.MustRegister(Transition{
		From: State("PENDING"),
		On:   Event("SUBMIT"),
		To:   State("SUBMITTED"),
	})

	next, _, err := m.Next("tx_1", State(" pending "), Event(" submit "), time.Now(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != State("SUBMITTED") {
		t.Fatalf("expected SUBMITTED got %s", next)
	}
}
