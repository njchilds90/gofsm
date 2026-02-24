package gofsm_test

import (
	"context"
	"fmt"

	"github.com/njchilds90/gofsm"
)

// ExampleFSM_Trigger demonstrates a basic traffic light state machine.
func ExampleFSM_Trigger() {
	f, _ := gofsm.New(gofsm.Config{
		Initial: "red",
		Transitions: []gofsm.Transition{
			{From: "red", Event: "go", To: "green"},
			{From: "green", Event: "slow", To: "yellow"},
			{From: "yellow", Event: "stop", To: "red"},
		},
	})

	ctx := context.Background()
	fmt.Println(f.Current())
	_ = f.Trigger(ctx, "go")
	fmt.Println(f.Current())
	_ = f.Trigger(ctx, "slow")
	fmt.Println(f.Current())

	// Output:
	// red
	// green
	// yellow
}

// ExampleFSM_Wildcard demonstrates using the Wildcard state to match any source state.
func ExampleFSM_Wildcard() {
	f, _ := gofsm.New(gofsm.Config{
		Initial: "running",
		Transitions: []gofsm.Transition{
			{From: "idle", Event: "start", To: "running"},
			{From: gofsm.Wildcard, Event: "emergency_stop", To: "idle"},
		},
	})

	ctx := context.Background()
	_ = f.Trigger(ctx, "emergency_stop")
	fmt.Println(f.Current())

	// Output:
	// idle
}

// ExampleFSM_Dot demonstrates generating a Graphviz DOT diagram.
func ExampleFSM_Dot() {
	f, _ := gofsm.New(gofsm.Config{
		Initial: "locked",
		Transitions: []gofsm.Transition{
			{From: "locked", Event: "coin", To: "unlocked"},
			{From: "unlocked", Event: "push", To: "locked"},
		},
	})

	dot := f.Dot()
	fmt.Println(len(dot) > 0)

	// Output:
	// true
}
