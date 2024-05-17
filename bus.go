package dew

import (
	"context"
	"fmt"
)

// New creates an instance of the Command Bus.
func New() Bus {
	return newMux()
}

// HandlerFunc defines a function type that takes a context and a command, returning an error.
type HandlerFunc[T any] func(ctx context.Context, command *T) error

// Handle calls the function f(ctx, command).
func (f HandlerFunc[T]) Handle(ctx context.Context, command *T) error {
	return f(ctx, command)
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

type BusContext struct {
	ctx context.Context

	// mwsIdx is the index of the middleware chain for method execution.
	mwsIdx int

	// handler is the wrapped handler function.
	handler finalHandler
}

func NewContext() *BusContext {
	return &BusContext{}
}

// Command returns the command object to be processed.
func (c *BusContext) Command() Command {
	if c.handler == nil {
		return nil
	}
	return c.handler.Command()
}

// WithContext returns a new Context with the given context.
func (c *BusContext) WithContext(ctx context.Context) Context {
	c.ctx = ctx
	return c
}

func (c *BusContext) Copy(a *BusContext) *BusContext {
	c.ctx = a.ctx
	c.mwsIdx = a.mwsIdx
	c.handler = a.handler
	return c
}

func (c *BusContext) Reset() {
	c.ctx = nil
	c.mwsIdx = 0
	c.handler = nil
}

// Context returns the underlying context.Context.
// If no context is set, it returns context.Background().
func (c *BusContext) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// WithValue returns a new Context with the given key-value pair added to the context.
func (c *BusContext) WithValue(key, val any) Context {
	return c.WithContext(context.WithValue(c.ctx, key, val))
}

// MiddlewareFunc is a type adapter to convert a function to a Middleware.
type MiddlewareFunc func(ctx Context) error

// Handle calls the function h(ctx, command).
func (h MiddlewareFunc) Handle(ctx Context) error {
	return h(ctx)
}

// Middleware is an interface for handling middleware.
type Middleware interface {
	// Handle executes the middleware.
	Handle(ctx Context) error
}

// OpType represents the type of operation.
type OpType uint8

const (
	// ACTION indicates a command that modifies state.
	ACTION OpType = 1 << iota
	// QUERY indicates a command that fetches data.
	QUERY
)

const ALL OpType = ACTION | QUERY

var (
	// ErrValidationFailed is returned when the command validation fails.
	ErrValidationFailed = fmt.Errorf("validation failed")
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

var (
	_ Bus = (*Mux)(nil)
)
