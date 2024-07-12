package dew

import (
	"context"
)

// Bus contains the core methods for dispatching commands.
type Bus interface {
	// Register adds the handler to the mux for the given command type.
	// It finds the handler methods that have the following signature:
	//
	//	func (h *Handler) FooMethod(ctx context.Context, command *BarCommand) error
	Register(handler any)
	// Use appends the middlewares to the mux middleware chain.
	// The middleware chain will be executed in the order they were added.
	// These middlewares are executed per command instead of per dispatch / query.
	Use(op OpType, middlewares ...func(next Middleware) Middleware)
	// Group creates a new mux with a copy of the parent middlewares.
	Group(fn func(mx Bus)) Bus
	// UseDispatch appends the middlewares to the dispatch middleware chain.
	// Dispatch middlewares are executed only once per dispatch instead of per command.
	UseDispatch(middlewares ...func(next Middleware) Middleware)
	// UseQuery appends the middlewares to the query middleware chain.
	// Query middlewares are executed only once per query instead of per command.
	UseQuery(middlewares ...func(next Middleware) Middleware)
}

// FromContext returns the bus from the context.
func FromContext(ctx context.Context) Bus {
	return ctx.Value(busCtxKey{}).(Bus)
}

// Context represents the context for a command execution.
type Context interface {
	// Context returns the underlying context.Context.
	Context() context.Context
	// WithContext returns a new Context with the given context.
	WithContext(ctx context.Context) Context
	// WithValue returns a new Context with the given key-value pair added to the context.
	WithValue(key, val any) Context
	// Command returns the command object to be processed.
	Command() Command
}

// HandlerFunc defines a function type that takes a context and a command, returning an error.
type HandlerFunc[T any] func(ctx context.Context, command *T) error

// Handle calls the function f(ctx, command).
func (f HandlerFunc[T]) Handle(ctx context.Context, command *T) error {
	return f(ctx, command)
}

type busCtxKey struct{}
