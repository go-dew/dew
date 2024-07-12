package main

import (
	"context"
	"fmt"

	"github.com/go-dew/dew"
)

type HelloAction struct {
	Name string
}

func (c HelloAction) Validate(_ context.Context) error {
	if c.Name == "" {
		return fmt.Errorf("invalid name")
	}
	return nil
}

func main() {
	// Initialize the Command Bus.
	bus := dew.New()

	// Register handler for HelloArgs.
	bus.Register(dew.HandlerFunc[HelloAction](func(ctx context.Context, cmd *HelloAction) error {
		println(fmt.Sprintf("Hello, %s!", cmd.Name)) // Output: Hello, Dew!
		return nil
	}))

	// Dispatch HelloArgs.
	_ = dew.Dispatch(context.Background(), dew.NewAction(bus, &HelloAction{Name: "Dew"}))
}
