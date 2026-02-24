# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-24

### Added

- `FSM` type with full concurrency safety via `sync.RWMutex`
- `Config` struct for declarative FSM construction
- `Transition` struct with optional `Guard` function
- `Wildcard` (`"*"`) support in `From` state field
- `OnEnter`, `OnExit`, and `OnTransition` lifecycle hooks
- `Trigger(ctx, event)` — fires events with context propagation
- `Current()` — returns current state
- `Is(state)` — boolean state check
- `Can(event)` / `Cannot(event)` — transition availability checks
- `AvailableEvents()` — lists all fireable events from current state
- `AllStates()` — returns all unique states defined in the machine
- `History()` — returns copy of all completed transitions
- `Reset(state)` — restores state without hooks (for serialization)
- `Dot()` — Graphviz DOT diagram export
- Structured sentinel errors: `ErrEmptyInitialState`, `ErrNoTransitions`, `ErrInvalidTransition`, `ErrInvalidEvent`, `ErrGuardRejected`
- Full test suite with race detector
- GitHub Actions CI (Go 1.21, 1.22, 1.23)
- GoDoc-compatible package documentation and examples
