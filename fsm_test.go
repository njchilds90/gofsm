package gofsm_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/njchilds90/gofsm"
)

func trafficLight() (*gofsm.FSM, error) {
	return gofsm.New(gofsm.Config{
		Initial: "red",
		Transitions: []gofsm.Transition{
			{From: "red", Event: "go", To: "green"},
			{From: "green", Event: "slow", To: "yellow"},
			{From: "yellow", Event: "stop", To: "red"},
		},
	})
}

func TestNew_Valid(t *testing.T) {
	f, err := trafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Current() != "red" {
		t.Fatalf("expected initial state red, got %s", f.Current())
	}
}

func TestNew_EmptyInitial(t *testing.T) {
	_, err := gofsm.New(gofsm.Config{
		Initial: "",
		Transitions: []gofsm.Transition{
			{From: "a", Event: "x", To: "b"},
		},
	})
	if !errors.Is(err, gofsm.ErrEmptyInitialState) {
		t.Fatalf("expected ErrEmptyInitialState, got %v", err)
	}
}

func TestNew_NoTransitions(t *testing.T) {
	_, err := gofsm.New(gofsm.Config{Initial: "a"})
	if !errors.Is(err, gofsm.ErrNoTransitions) {
		t.Fatalf("expected ErrNoTransitions, got %v", err)
	}
}

func TestNew_DuplicateTransition(t *testing.T) {
	_, err := gofsm.New(gofsm.Config{
		Initial: "a",
		Transitions: []gofsm.Transition{
			{From: "a", Event: "go", To: "b"},
			{From: "a", Event: "go", To: "c"},
		},
	})
	if !errors.Is(err, gofsm.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestNew_EmptyFrom(t *testing.T) {
	_, err := gofsm.New(gofsm.Config{
		Initial: "a",
		Transitions: []gofsm.Transition{
			{From: "", Event: "go", To: "b"},
		},
	})
	if !errors.Is(err, gofsm.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestNew_EmptyEvent(t *testing.T) {
	_, err := gofsm.New(gofsm.Config{
		Initial: "a",
		Transitions: []gofsm.Transition{
			{From: "a", Event: "", To: "b"},
		},
	})
	if !errors.Is(err, gofsm.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestNew_EmptyTo(t *testing.T) {
	_, err := gofsm.New(gofsm.Config{
		Initial: "a",
		Transitions: []gofsm.Transition{
			{From: "a", Event: "go", To: ""},
		},
	})
	if !errors.Is(err, gofsm.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestTrigger_Valid(t *testing.T) {
	f, _ := trafficLight()
	ctx := context.Background()

	if err := f.Trigger(ctx, "go"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Current() != "green" {
		t.Fatalf("expected green, got %s", f.Current())
	}
}

func TestTrigger_InvalidEvent(t *testing.T) {
	f, _ := trafficLight()
	err := f.Trigger(context.Background(), "slow")
	if !errors.Is(err, gofsm.ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestTrigger_GuardRejects(t *testing.T) {
	guardErr := errors.New("not authorized")
	f, _ := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{
				From:  "idle",
				Event: "start",
				To:    "running",
				Guard: func(ctx context.Context, from, to, event string) error {
					return guardErr
				},
			},
		},
	})
	err := f.Trigger(context.Background(), "start")
	if !errors.Is(err, gofsm.ErrGuardRejected) {
		t.Fatalf("expected ErrGuardRejected, got %v", err)
	}
}

func TestTrigger_GuardAllows(t *testing.T) {
	f, _ := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{
				From:  "idle",
				Event: "start",
				To:    "running",
				Guard: func(ctx context.Context, from, to, event string) error {
					return nil
				},
			},
		},
	})
	if err := f.Trigger(context.Background(), "start"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Current() != "running" {
		t.Fatalf("expected running, got %s", f.Current())
	}
}

func TestIs(t *testing.T) {
	f, _ := trafficLight()
	if !f.Is("red") {
		t.Fatal("expected Is(red) to be true")
	}
	if f.Is("green") {
		t.Fatal("expected Is(green) to be false")
	}
}

func TestCan_Cannot(t *testing.T) {
	f, _ := trafficLight()
	if !f.Can("go") {
		t.Fatal("expected Can(go) to be true")
	}
	if !f.Cannot("slow") {
		t.Fatal("expected Cannot(slow) to be true")
	}
}

func TestHistory(t *testing.T) {
	f, _ := trafficLight()
	ctx := context.Background()
	_ = f.Trigger(ctx, "go")
	_ = f.Trigger(ctx, "slow")

	h := f.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(h))
	}
	if h[0].From != "red" || h[0].Event != "go" || h[0].To != "green" {
		t.Fatalf("unexpected first entry: %+v", h[0])
	}
	if h[1].From != "green" || h[1].Event != "slow" || h[1].To != "yellow" {
		t.Fatalf("unexpected second entry: %+v", h[1])
	}
}

func TestReset(t *testing.T) {
	f, _ := trafficLight()
	f.Reset("yellow")
	if f.Current() != "yellow" {
		t.Fatalf("expected yellow after reset, got %s", f.Current())
	}
}

func TestAllStates(t *testing.T) {
	f, _ := trafficLight()
	states := f.AllStates()
	seen := make(map[string]bool)
	for _, s := range states {
		seen[s] = true
	}
	for _, want := range []string{"red", "green", "yellow"} {
		if !seen[want] {
			t.Fatalf("expected state %q in AllStates", want)
		}
	}
}

func TestAvailableEvents(t *testing.T) {
	f, _ := trafficLight()
	events := f.AvailableEvents()
	if len(events) != 1 || events[0] != "go" {
		t.Fatalf("expected [go], got %v", events)
	}
}

func TestWildcard(t *testing.T) {
	f, err := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{From: "idle", Event: "start", To: "running"},
			{From: gofsm.Wildcard, Event: "reset", To: "idle"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	_ = f.Trigger(ctx, "start")
	if f.Current() != "running" {
		t.Fatalf("expected running, got %s", f.Current())
	}
	if err := f.Trigger(ctx, "reset"); err != nil {
		t.Fatalf("unexpected error on wildcard: %v", err)
	}
	if f.Current() != "idle" {
		t.Fatalf("expected idle after reset, got %s", f.Current())
	}
}

func TestOnEnterOnExitHooks(t *testing.T) {
	var log []string
	f, _ := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{From: "idle", Event: "start", To: "running"},
		},
		OnExit: map[string]gofsm.HookFunc{
			"idle": func(ctx context.Context, from, to, event string) {
				log = append(log, "exit:idle")
			},
		},
		OnEnter: map[string]gofsm.HookFunc{
			"running": func(ctx context.Context, from, to, event string) {
				log = append(log, "enter:running")
			},
		},
	})
	_ = f.Trigger(context.Background(), "start")
	if len(log) != 2 || log[0] != "exit:idle" || log[1] != "enter:running" {
		t.Fatalf("unexpected hook log: %v", log)
	}
}

func TestOnTransitionHook(t *testing.T) {
	var called bool
	f, _ := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{From: "idle", Event: "start", To: "running"},
		},
		OnTransition: func(ctx context.Context, from, to, event string) {
			called = true
		},
	})
	_ = f.Trigger(context.Background(), "start")
	if !called {
		t.Fatal("OnTransition hook was not called")
	}
}

func TestDot(t *testing.T) {
	f, _ := trafficLight()
	dot := f.Dot()
	if !strings.Contains(dot, "digraph FSM") {
		t.Fatal("Dot output missing digraph header")
	}
	if !strings.Contains(dot, "red") {
		t.Fatal("Dot output missing state 'red'")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f, _ := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{
				From:  "idle",
				Event: "start",
				To:    "running",
				Guard: func(ctx context.Context, from, to, event string) error {
					return ctx.Err()
				},
			},
		},
	})
	err := f.Trigger(ctx, "start")
	if !errors.Is(err, gofsm.ErrGuardRejected) {
		t.Fatalf("expected ErrGuardRejected wrapping context error, got %v", err)
	}
}

func TestConcurrentTrigger(t *testing.T) {
	f, _ := gofsm.New(gofsm.Config{
		Initial: "idle",
		Transitions: []gofsm.Transition{
			{From: "idle", Event: "start", To: "running"},
			{From: "running", Event: "stop", To: "idle"},
		},
	})

	done := make(chan struct{})
	go func() {
		ctx := context.Background()
		for i := 0; i < 100; i++ {
			_ = f.Trigger(ctx, "start")
			_ = f.Trigger(ctx, "stop")
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		_ = f.Current()
		_ = f.Can("start")
	}
	<-done
}
