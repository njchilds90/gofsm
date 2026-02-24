# Contributing to gofsm

Thank you for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass: `go test -race ./...`
6. Run `go vet ./...` and `go fmt ./...`
7. Open a pull request against `main`

## Guidelines

- Keep the zero-dependency policy — do not add external imports
- All public functions must have GoDoc comments
- All new behavior must have table-driven tests
- Maintain backward compatibility within a major version
- Follow idiomatic Go conventions

## Reporting Issues

Open a GitHub Issue with a clear description, reproduction steps, and Go version.
```

---

---
## gofsm v1.0.0

First stable release of `gofsm` — a clean, deterministic finite state machine library for Go.

### Features
- Declarative FSM construction via `Config` struct
- Guard functions to cancel transitions
- `OnEnter`, `OnExit`, and `OnTransition` lifecycle hooks
- Wildcard (`*`) transitions matching any source state
- Full transition history tracking
- `Dot()` Graphviz export for visualization
- `Reset()` for state restoration without hooks
- Structured sentinel errors compatible with `errors.Is`
- Concurrency-safe via `sync.RWMutex`
- Zero external dependencies

### Install
```
go get github.com/njchilds90/gofsm@v1.0.0
```
---

7. Leave "Set as the latest release" checked ✅

8. Click the green "Publish release" button.

pkg.go.dev will auto-index within ~10 minutes.

Verify at:
https://pkg.go.dev/github.com/njchilds90/gofsm
