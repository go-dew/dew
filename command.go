package dew

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"
)

// Command represents an Action or QueryAction.
type Command interface{}

// Action represents a mutable action.
type Action interface {
	// Validate validates the command.
	Validate(context.Context) error
}

// QueryAction represents a read-only action.
type QueryAction interface{}

// Commands is a collection of CommandHandlers.
type Commands []CommandHandler[Command]

// Actions is a collection of CommandHandlers.
type Actions []CommandHandler[Action]

// CommandHandler represents a command to be dispatched.
type CommandHandler[T Command] interface {
	Handle(ctx Context) error
	Command() Command
	Mux() *Mux
}

// NewAction creates an object that can be dispatched.
// It panics if the handler is not found.
func NewAction[T Action](bus Bus, cmd *T) CommandHandler[T] {
	h, mx := resolveHandler[T](ACTION, bus)
	return command[T]{
		mux:     mx,
		cmd:     cmd,
		handler: h,
	}
}

// NewQuery creates an object that can be dispatched.
// It panics if the handler is not found.
func NewQuery[T QueryAction](bus Bus, cmd *T) CommandHandler[T] {
	h, mx := resolveHandler[T](QUERY, bus)
	return command[T]{
		mux:     mx,
		cmd:     cmd,
		handler: h,
	}
}

// command carries the necessary information to dispatch a command.
type command[T Command] struct {
	mux     *Mux
	cmd     *T
	handler HandlerFunc[T]
}

func (c command[T]) Handle(ctx Context) error {
	return c.handler(ctx.Context(), c.cmd)
}

func (c command[T]) Command() Command {
	return c.cmd
}

func (c command[T]) Mux() *Mux {
	return c.mux
}

// resolveHandler locates a handler for a given operation type and command type within the provided Bus instance.
// It constructs a key from the command's reflect.Type, then searches the Mux's tree structure for a corresponding node.
//
// Parameters:
// - typ: The reflect.Type of the command for which a handler is being sought.
// - op: The operation type (ACTION or QUERY) under which the handler should be classified.
// - bus: The Bus instance where handlers are registered and organized.
//
// Returns:
// - *node: A pointer to the node struct representing the handler if found.
// - error: An error if no handler could be found for the provided type and operation.
//
// Example:
//
//	handlerNode, err := resolveHandler(reflect.TypeOf(myCommand), ACTION, myBus)
//	if err != nil {
//	  log.Fatalf("Handler resolution failed: %v", err)
//	}
func convertInterface[T any](i any) T {
	var v T
	vp := unsafe.Pointer(&v)
	reflect.NewAt(reflect.TypeOf(v), vp).Elem().Set(reflect.ValueOf(i))
	return v
}

// resolveHandler returns the handler and mux for the given command.
func resolveHandler[T Command](op OpType, bus Bus) (HandlerFunc[T], *Mux) {
	k := getKey[T]()
	mx := bus.(*Mux)
	n := mx.tree.findRoute(op, k)
	if n != nil {
		h := n.handler.handler
		return convertInterface[HandlerFunc[T]](h.handler), h.mux
	}
	panic(fmt.Sprintf("handler not found for %s", k))
}
