# gofsm

A clean, deterministic finite state machine (FSM) library for Go — with guards, lifecycle hooks, context propagation, wildcard transitions, history tracking, and Graphviz DOT export.

[![CI](https://github.com/njchilds90/gofsm/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/gofsm/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/gofsm.svg)](https://pkg.go.dev/github.com/njchilds90/gofsm)
[![Go Report Card](https://goreportcard.com/badge/github.com/njchilds90/gofsm)](https://goreportcard.com/report/github.com/njchilds90/gofsm)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Why gofsm?

Python has [`transitions`](https://github.com/pytransitions/transitions) (4k+ stars). JavaScript has [`XState`](https://github.com/statelyai/xstate) (27k+ stars). Rust has `sm`. Go has no clean, idiomatic, zero-dependency FSM library.

`gofsm` fills that gap. It is:

- **Zero dependencies** — only the Go standard library
- **Concurrency safe** — `sync.RWMutex` protected
- **AI-agent friendly** — deterministic, inspectable, composable
- **Context-aware** — all hooks and guards receive `context.Context`
- **Fully inspectable** — history, available events, DOT export

## Installation
```bash
go get github.com/njchilds90/gofsm@v1.0.0
```

## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "github.com/njchilds90/gofsm"
)

func main() {
    f, err := gofsm.New(gofsm.Config{
        Initial: "idle",
        Transitions: []gofsm.Transition{
            {From: "idle",    Event: "start", To: "running"},
            {From: "running", Event: "pause", To: "paused"},
            {From: "paused",  Event: "resume", To: "running"},
            {From: "running", Event: "stop",  To: "idle"},
        },
    })
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    fmt.Println(f.Current()) // idle

    _ = f.Trigger(ctx, "start")
    fmt.Println(f.Current()) // running

    _ = f.Trigger(ctx, "pause")
    fmt.Println(f.Current()) // paused
}
```

## Features

### Guards

Guards are functions that can cancel a transition before it fires.
```go
gofsm.Transition{
    From:  "idle",
    Event: "start",
    To:    "running",
    Guard: func(ctx context.Context, from, to, event string) error {
        if !isAuthorized(ctx) {
            return errors.New("not authorized")
        }
        return nil
    },
}
```

### Lifecycle Hooks
```go
gofsm.Config{
    OnEnter: map[string]gofsm.HookFunc{
        "running": func(ctx context.Context, from, to, event string) {
            fmt.Println("entered running state")
        },
    },
    OnExit: map[string]gofsm.HookFunc{
        "idle": func(ctx context.Context, from, to, event string) {
            fmt.Println("leaving idle state")
        },
    },
    OnTransition: func(ctx context.Context, from, to, event string) {
        fmt.Printf("transition: %s -[%s]-> %s\n", from, event, to)
    },
}
```

### Wildcard Transitions

Use `gofsm.Wildcard` (`"*"`) in `From` to match any state:
```go
{From: gofsm.Wildcard, Event: "emergency_stop", To: "idle"}
```

### Inspection
```go
f.Current()          // current state
f.Is("running")      // bool
f.Can("stop")        // bool
f.Cannot("start")    // bool
f.AvailableEvents()  // []string
f.AllStates()        // []string
f.History()          // []HistoryEntry
f.Dot()              // Graphviz DOT string
```

### Reset (State Restoration)
```go
f.Reset("paused") // move to any state without firing hooks
```

## Error Handling

All errors are sentinel values comparable with `errors.Is`:

| Error | Cause |
|-------|-------|
| `ErrEmptyInitialState` | `Config.Initial` is empty |
| `ErrNoTransitions` | `Config.Transitions` is empty |
| `ErrInvalidTransition` | Malformed transition (empty field or duplicate) |
| `ErrInvalidEvent` | Event fired with no matching transition |
| `ErrGuardRejected` | Guard function returned an error |

## Design Philosophy

- **Pure functions where possible** — no hidden global state
- **Deterministic** — same inputs always produce same outputs
- **Minimal but complete** — small API surface, every method earns its place
- **AI-agent composable** — structured errors, context propagation, inspectable state

## License

MIT