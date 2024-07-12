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
}

// NewAction creates an object that can be dispatched.
// It panics if the handler is not found.
func NewAction[T Action](bus Bus, cmd *T) CommandHandler[T] {
	h, mx := resolveHandler[T](bus)
	return command[T]{
		mux:     mx,
		cmd:     cmd,
		handler: h,
	}
}

// NewQuery creates an object that can be dispatched.
// It panics if the handler is not found.
func NewQuery[T QueryAction](bus Bus, cmd *T) CommandHandler[T] {
	h, mx := resolveHandler[T](bus)
	return command[T]{
		mux:     mx,
		cmd:     cmd,
		handler: h,
	}
}

// command carries the necessary information to dispatch a command.
type command[T Command] struct {
	mux     *mux
	cmd     *T
	handler HandlerFunc[T]
}

func (c command[T]) Handle(ctx Context) error {
	return c.handler(ctx.Context(), c.cmd)
}

func (c command[T]) Command() Command {
	return c.cmd
}

func (c command[T]) Mux() *mux {
	return c.mux
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

// resolveHandler returns the handler and mux for the given command.
func resolveHandler[T Command](bus Bus) (HandlerFunc[T], *mux) {
	typ := typeFor[T]()
	mx := bus.(*mux)

	h, mxx, ok := loadHandlerCache[T](typ, mx)
	if ok {
		return h, mxx
	}

	entry, ok := mx.entries.Load(typ)
	if ok {
		hh := entry.(*handler)
		hhh := convertInterface[HandlerFunc[T]](hh.handler)
		storeCache[T](mx.cache, typ, hh.mux, hhh)
		return hhh, hh.mux
	}

	panic(fmt.Sprintf("handler not found for %s/%s", typ.PkgPath(), typ.String()))
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
