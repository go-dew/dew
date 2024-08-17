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
	Mux() *mux
	Resolve(bus Bus) error
}

// NewAction creates an object that can be dispatched.
// It panics if the handler is not found.
func NewAction[T Action](cmd *T) CommandHandler[T] {
	typ := typeFor[T]()
	return &command[T]{
		cmd: cmd,
		typ: typ,
	}
}

// NewQuery creates an object that can be dispatched.
// It panics if the handler is not found.
func NewQuery[T QueryAction](cmd *T) CommandHandler[T] {
	typ := typeFor[T]()
	return &command[T]{
		cmd: cmd,
		typ: typ,
	}
}

// command carries the necessary information to dispatch a command.
type command[T Command] struct {
	mux     *mux
	cmd     *T
	handler HandlerFunc[T]
	typ     reflect.Type
}

func (c *command[T]) Handle(ctx Context) error {
	return c.handler(ctx.Context(), c.cmd)
}

func (c *command[T]) Command() Command {
	return c.cmd
}

func (c *command[T]) Mux() *mux {
	return c.mux
}

func (c *command[T]) Resolve(bus Bus) error {
	mx := bus.(*mux)

	h, mxx, ok := loadHandlerCache[T](c.typ, mx)
	if ok {
		c.handler = h
		c.mux = mxx
		return nil
	}

	entry, ok := mx.entries.Load(c.typ)
	if ok {
		hh := entry.(*handler)
		hhh := convertInterface[HandlerFunc[T]](hh.handler)
		storeCache[T](mx.cache, c.typ, hh.mux, hhh)
		c.handler = hhh
		c.mux = hh.mux
		return nil
	}

	return fmt.Errorf("handler not found for %v", c.typ)
}

func convertInterface[T any](i any) T {
	var v T
	reflect.NewAt(reflect.TypeOf(v), unsafe.Pointer(&v)).Elem().Set(reflect.ValueOf(i))
	return v
}

type entry struct {
	t reflect.Type
	p unsafe.Pointer
	m *mux
}

// storeCache stores the handler in the cache.
func storeCache[T Command](cache *syncMap, t reflect.Type, mx *mux, handlerFunc HandlerFunc[T]) {
	cache.store(t, entry{t: t, m: mx, p: unsafe.Pointer(&handlerFunc)})
}

// loadHandlerCache loads the handler from the cache.
func loadHandlerCache[T Command](typ reflect.Type, mx *mux) (HandlerFunc[T], *mux, bool) {
	if v, ok := mx.cache.load(typ); ok {
		e := v.(entry)
		return *(*HandlerFunc[T])(e.p), e.m, true
	}
	return nil, nil, false
}

// typeFor returns the reflect.Type for the given type.
func typeFor[T any]() reflect.Type {
	var t T
	return reflect.TypeOf(t)
}

type handler struct {
	// handler is the function to call.
	handler any
	// mux is the mux that the handler belongs to.
	mux *mux
}
