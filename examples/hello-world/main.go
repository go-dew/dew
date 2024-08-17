package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-dew/dew"
)

// HelloAction represents the action of greeting someone.
type HelloAction struct {
	Name string
}

// Validate checks if the HelloAction is valid.
func (c HelloAction) Validate(_ context.Context) error {
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

// HelloHandler handles the HelloAction.
type HelloHandler struct{}

// HandleHello is the handler function for HelloAction.
func (h *HelloHandler) HandleHello(ctx context.Context, cmd *HelloAction) error {
	fmt.Printf("Hello, %s!\n", cmd.Name)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	// Initialize the Command Bus.
	bus := dew.New()

	// Register HelloHandler.
	bus.Register(&HelloHandler{})

	// Create a context with the bus.
	ctx := dew.NewContext(context.Background(), bus)

	// Get the name from command-line arguments or use a default.
	name := "Dew"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	// Create and dispatch HelloAction.
	action := &HelloAction{Name: name}
	if err := dew.Dispatch(ctx, dew.NewAction(action)); err != nil {
		return fmt.Errorf("failed to dispatch HelloAction: %w", err)
	}

	return nil
}
