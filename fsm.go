// Package gofsm provides a clean, deterministic finite state machine (FSM)
// for Go with support for guards, lifecycle hooks, context propagation,
// and structured transition definitions.
//
// It is designed to be composable, AI-agent-friendly, and free of global state.
//
// Basic usage:
//
//	f, err := gofsm.New(gofsm.Config{
//	    Initial: "idle",
//	    Transitions: []gofsm.Transition{
//	        {From: "idle", Event: "start", To: "running"},
//	        {From: "running", Event: "pause", To: "paused"},
//	        {From: "paused", Event: "resume", To: "running"},
//	        {From: "running", Event: "stop", To: "idle"},
//	    },
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = f.Trigger(context.Background(), "start")
package gofsm

import (
	"context"
	"fmt"
	"sync"
)

// State represents a named FSM state.
type State = string

// Event represents a named FSM trigger.
type Event = string

// GuardFunc is called before a transition fires. Return a non-nil error to cancel the transition.
type GuardFunc func(ctx context.Context, from, to State, event Event) error

// HookFunc is called after a transition fires successfully.
type HookFunc func(ctx context.Context, from, to State, event Event)

// Transition defines a valid state change in the FSM.
type Transition struct {
	// From is the source state. Use Wildcard ("*") to match any state.
	From State
	// Event is the trigger that causes this transition.
	Event Event
	// To is the destination state.
	To State
	// Guard is an optional function called before the transition.
	// If Guard returns an error, the transition is cancelled.
	Guard GuardFunc
}

// Config holds the configuration used to build an FSM.
type Config struct {
	// Initial is the starting state of the FSM.
	Initial State
	// Transitions is the list of valid transitions.
	Transitions []Transition
	// OnEnter hooks are called after entering a state.
	// Map key is the state name; use Wildcard ("*") to match all states.
	OnEnter map[State]HookFunc
	// OnExit hooks are called before leaving a state.
	// Map key is the state name; use Wildcard ("*") to match all states.
	OnExit map[State]HookFunc
	// OnTransition is called after every successful transition.
	OnTransition HookFunc
}

// Wildcard matches any state in From fields and in hook maps.
const Wildcard State = "*"

// FSM is a finite state machine. It is safe for concurrent use.
type FSM struct {
	mu          sync.RWMutex
	current     State
	transitions map[transitionKey]Transition
	onEnter     map[State]HookFunc
	onExit      map[State]HookFunc
	onTransition HookFunc
	history     []HistoryEntry
}

type transitionKey struct {
	from  State
	event Event
}

// HistoryEntry records a completed transition.
type HistoryEntry struct {
	From  State
	Event Event
	To    State
}

// New creates and validates a new FSM from the provided Config.
// Returns an error if the config is invalid (e.g. empty initial state,
// duplicate transitions, or empty event/state names).
func New(cfg Config) (*FSM, error) {
	if cfg.Initial == "" {
		return nil, ErrEmptyInitialState
	}
	if len(cfg.Transitions) == 0 {
		return nil, ErrNoTransitions
	}

	tmap := make(map[transitionKey]Transition, len(cfg.Transitions))
	for _, t := range cfg.Transitions {
		if t.From == "" {
			return nil, fmt.Errorf("%w: transition has empty From state", ErrInvalidTransition)
		}
		if t.Event == "" {
			return nil, fmt.Errorf("%w: transition has empty Event", ErrInvalidTransition)
		}
		if t.To == "" {
			return nil, fmt.Errorf("%w: transition has empty To state", ErrInvalidTransition)
		}
		key := transitionKey{from: t.From, event: t.Event}
		if _, exists := tmap[key]; exists {
			return nil, fmt.Errorf("%w: duplicate transition from=%q event=%q", ErrInvalidTransition, t.From, t.Event)
		}
		tmap[key] = t
	}

	onEnter := cfg.OnEnter
	if onEnter == nil {
		onEnter = make(map[State]HookFunc)
	}
	onExit := cfg.OnExit
	if onExit == nil {
		onExit = make(map[State]HookFunc)
	}

	return &FSM{
		current:      cfg.Initial,
		transitions:  tmap,
		onEnter:      onEnter,
		onExit:       onExit,
		onTransition: cfg.OnTransition,
		history:      []HistoryEntry{},
	}, nil
}

// Current returns the current state of the FSM.
func (f *FSM) Current() State {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}

// Is returns true if the FSM is in the given state.
func (f *FSM) Is(state State) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current == state
}

// Can returns true if the given event can fire from the current state.
func (f *FSM) Can(event Event) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.findTransition(f.current, event)
	return ok
}

// Cannot returns true if the given event cannot fire from the current state.
func (f *FSM) Cannot(event Event) bool {
	return !f.Can(event)
}

// Trigger fires an event, transitioning the FSM if valid.
// Returns ErrInvalidEvent if no matching transition exists.
// Returns ErrGuardRejected (wrapping the guard's error) if a guard function rejects the transition.
// Trigger is safe to call concurrently.
func (f *FSM) Trigger(ctx context.Context, event Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.findTransition(f.current, event)
	if !ok {
		return fmt.Errorf("%w: event=%q state=%q", ErrInvalidEvent, event, f.current)
	}

	if t.Guard != nil {
		if err := t.Guard(ctx, f.current, t.To, event); err != nil {
			return fmt.Errorf("%w: %v", ErrGuardRejected, err)
		}
	}

	from := f.current

	// OnExit hooks
	if fn, ok := f.onExit[from]; ok {
		fn(ctx, from, t.To, event)
	}
	if fn, ok := f.onExit[Wildcard]; ok {
		fn(ctx, from, t.To, event)
	}

	f.current = t.To

	// OnEnter hooks
	if fn, ok := f.onEnter[t.To]; ok {
		fn(ctx, from, t.To, event)
	}
	if fn, ok := f.onEnter[Wildcard]; ok {
		fn(ctx, from, t.To, event)
	}

	// OnTransition hook
	if f.onTransition != nil {
		f.onTransition(ctx, from, t.To, event)
	}

	f.history = append(f.history, HistoryEntry{From: from, Event: event, To: t.To})

	return nil
}

// AvailableEvents returns the list of events that can fire from the current state.
func (f *FSM) AvailableEvents() []Event {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var events []Event
	for key := range f.transitions {
		if key.from == f.current || key.from == Wildcard {
			events = append(events, key.event)
		}
	}
	return events
}

// AllStates returns all unique states defined in the FSM (From and To).
func (f *FSM) AllStates() []State {
	f.mu.RLock()
	defer f.mu.RUnlock()

	seen := make(map[State]struct{})
	for key, t := range f.transitions {
		if key.from != Wildcard {
			seen[key.from] = struct{}{}
		}
		seen[t.To] = struct{}{}
	}
	states := make([]State, 0, len(seen))
	for s := range seen {
		states = append(states, s)
	}
	return states
}

// History returns a copy of all completed transitions since the FSM was created.
func (f *FSM) History() []HistoryEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()
	cp := make([]HistoryEntry, len(f.history))
	copy(cp, f.history)
	return cp
}

// Reset moves the FSM to the given state without triggering any hooks or guards.
// This is useful for restoring serialized state.
func (f *FSM) Reset(state State) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.current = state
}

// Dot returns a Graphviz DOT representation of the FSM for visualization.
func (f *FSM) Dot() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := "digraph FSM {\n\trankdir=LR;\n"
	out += fmt.Sprintf("\t\"%s\" [shape=doublecircle];\n", f.current)
	for key, t := range f.transitions {
		from := key.from
		if from == Wildcard {
			from = "*"
		}
		out += fmt.Sprintf("\t\"%s\" -> \"%s\" [label=\"%s\"];\n", from, t.To, key.event)
	}
	out += "}\n"
	return out
}

// findTransition looks up a transition by current state and event.
// It prefers exact matches over wildcard matches.
// Must be called with f.mu held (at least read lock).
func (f *FSM) findTransition(state State, event Event) (Transition, bool) {
	if t, ok := f.transitions[transitionKey{from: state, event: event}]; ok {
		return t, true
	}
	if t, ok := f.transitions[transitionKey{from: Wildcard, event: event}]; ok {
		return t, true
	}
	return Transition{}, false
}
