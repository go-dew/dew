package dew

import "context"

type BusContext struct {
	ctx context.Context

	// mwsIdx is the index of the middleware chain for method execution.
	mwsIdx int

	// handler is the wrapped handler function.
	handler internalHandler
}

func NewContext() *BusContext {
	return &BusContext{}
}

type internalHandler interface {
	Handle(ctx Context) error
	Command() Command
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
