package gofsm

import "errors"

// Sentinel errors returned by the FSM. All errors are safe to compare with errors.Is.

// ErrEmptyInitialState is returned when Config.Initial is empty.
var ErrEmptyInitialState = errors.New("gofsm: initial state must not be empty")

// ErrNoTransitions is returned when Config.Transitions is empty.
var ErrNoTransitions = errors.New("gofsm: at least one transition must be defined")

// ErrInvalidTransition is returned when a transition definition is malformed.
var ErrInvalidTransition = errors.New("gofsm: invalid transition definition")

// ErrInvalidEvent is returned when Trigger is called with an event that has no
// matching transition from the current state.
var ErrInvalidEvent = errors.New("gofsm: no valid transition for event in current state")

// ErrGuardRejected is returned when a guard function cancels a transition.
var ErrGuardRejected = errors.New("gofsm: transition rejected by guard")
