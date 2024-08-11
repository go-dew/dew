package dew

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

var (
	_ Bus = (*mux)(nil)
)

// mux is the main struct where all handlers and middlewares are registered.
type mux struct {
	parent      *mux
	inline      bool
	lock        sync.RWMutex
	entries     *sync.Map
	handler     [ALL]Middleware
	middlewares [mAll][]middleware
	mHandlers   [mAll]func(ctx Context, fn mHandlerFunc) error
	cache       *syncMap

	// context pool
	pool *sync.Pool
}

// New creates an instance of the Command Bus.
func New() Bus {
	return newMux()
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

type mHandlerFunc func(ctx Context) error

type middlewareType int

const (
	mCmd middlewareType = iota
	mDispatch
	mQuery
	mAll
)

// newMux returns a newly initialized Mux object that implements the dispatcher interface.
func newMux() *mux {
	mux := &mux{entries: &sync.Map{}, pool: &sync.Pool{}}
	mux.pool.New = func() interface{} {
		return NewContext()
	}
	mux.cache = &syncMap{kv: make(map[reflect.Type]any)}
	return mux
}

func (mx *mux) root() *mux {
	if mx.parent == nil {
		return mx
	}
	return mx.parent.root()
}

// Use appends the middlewares to the mux middleware chain.
// The middleware chain will be executed in the order they were added.
func (mx *mux) Use(op OpType, middlewares ...func(next Middleware) Middleware) {
	for _, mw := range middlewares {
		mx.middlewares[mCmd] = append(mx.middlewares[mCmd], middleware{op: op, fn: mw})
	}
}

// UseDispatch appends the middlewares to the dispatch middleware chain.
func (mx *mux) UseDispatch(middlewares ...func(next Middleware) Middleware) {
	mx.addMiddleware(mDispatch, middlewares)
}

// UseQuery appends the middlewares to the query middleware chain.
func (mx *mux) UseQuery(middlewares ...func(next Middleware) Middleware) {
	mx.addMiddleware(mQuery, middlewares)
}

func (mx *mux) addMiddleware(m middlewareType, mws []func(next Middleware) Middleware) {
	for _, mw := range mws {
		mx.middlewares[m] = append(mx.middlewares[m], middleware{fn: mw})
	}
}

// Group creates a new mux with a copy of the parent middlewares.
func (mx *mux) Group(fn func(mx Bus)) Bus {
	child := mx.child()
	if fn != nil {
		fn(child)
	}
	return child
}

// with creates a new mux with the given middlewares.
func (mx *mux) child() Bus {

	// copy the parent middlewares
	var mws [mAll][]middleware
	for i := range mws {
		mws[i] = make([]middleware, len(mx.middlewares[i]))
		copy(mws[i], mx.middlewares[i])
	}

	return &mux{
		parent:      mx,
		inline:      true,
		middlewares: mws,
		entries:     mx.entries,
		cache:       mx.cache,
	}
}

// dispatch dispatches the command to the appropriate Executor.
func (mx *mux) dispatch(op OpType, ctx Context, h internalHandler) error {
	hh := mx.handlerFor(op)
	if hh == nil {
		mx.updateRouteHandler(op)
		hh = mx.handlerFor(op)
	}
	ctx.(*BusContext).handler = h
	return hh.Handle(ctx)
}

func (mx *mux) handlerFor(op OpType) Middleware {
	mx.lock.RLock()
	defer mx.lock.RUnlock()
	return mx.handler[op]
}

func (mx *mux) newDispatchHandler(m middlewareType, fn func(ctx Context) error) Middleware {
	return exec(mx.middlewares[m], MiddlewareFunc(
		func(ctx Context) error {
			return fn(ctx)
		}))
}

func (mx *mux) updateRouteHandler(op OpType) {
	mx.lock.Lock()
	defer mx.lock.Unlock()
	mx.handler[op] = chain(op, mx.middlewares[mCmd], MiddlewareFunc(
		func(ctx Context) error {
			return ctx.(*BusContext).handler.Handle(ctx)
		}))
}

func (mx *mux) updateHandler(m middlewareType) {
	mx.lock.Lock()
	defer mx.lock.Unlock()
	mx.mHandlers[m] = func(ctx Context, fn mHandlerFunc) error {
		return mx.newDispatchHandler(m, func(ctx Context) error {
			return fn(ctx)
		}).Handle(ctx)
	}
}

// Register adds the handler to the mux for the given command type.
func (mx *mux) Register(handler any) {
	hdlTyp := reflect.TypeOf(handler)

	// if it's not a pointer, convert it to a pointer so we can get
	// access to all the methods of the struct.
	if hdlTyp.Kind() != reflect.Ptr {
		newVal := reflect.New(reflect.PointerTo(hdlTyp))
		ptrToHandler := reflect.ValueOf(&handler)
		if ptrToHandler.Elem().Type().AssignableTo(newVal.Elem().Type()) {
			newVal.Elem().Set(ptrToHandler.Elem())
		} else {
			// If direct assignment is not possible, we need to create an intermediate value
			intermediateVal := reflect.New(hdlTyp)
			intermediateVal.Elem().Set(reflect.ValueOf(handler))
			newVal.Elem().Set(intermediateVal)
		}
		handler = newVal.Interface()
		hdlTyp = reflect.TypeOf(handler)

		if hdlTyp.Kind() == reflect.Ptr && hdlTyp.Elem().Kind() == reflect.Ptr {
			// We have a pointer to a pointer, let's get to the single pointer
			newVal := reflect.ValueOf(handler).Elem()
			handler = newVal.Interface()
			hdlTyp = reflect.TypeOf(handler)
		}

	}

	kind := hdlTyp.Kind()
	println(fmt.Sprintf("Register: %v %s", kind, hdlTyp.String()))

	numMethod := hdlTyp.NumMethod()
	println(fmt.Sprintf("numMethod: %v", numMethod))

	for i := 0; i < hdlTyp.NumMethod(); i++ {
		mtdTyp := hdlTyp.Method(i)
		if isHandlerMethod(mtdTyp) {
			cmdTyp := mtdTyp.Type.In(2).Elem()
			if cmdTyp.Implements(reflect.TypeOf((*Action)(nil)).Elem()) {
				mx.addHandler(cmdTyp, reflect.ValueOf(handler).Method(i).Interface())
			} else if cmdTyp.Implements(reflect.TypeOf((*QueryAction)(nil)).Elem()) {
				mx.addHandler(cmdTyp, reflect.ValueOf(handler).Method(i).Interface())
			}
		}
	}
	mx.setupHandler()
}

func (mx *mux) setupHandler() {
	if mx.mHandlers[mQuery] == nil {
		mx.updateHandler(mQuery)
	}
	if mx.mHandlers[mDispatch] == nil {
		println("setupHandler")
		mx.updateHandler(mDispatch)
	}
	if mx.parent != nil {
		mx.parent.setupHandler()
	}
}

func (mx *mux) addHandler(t reflect.Type, h any) {
	println(fmt.Sprintf("addHandler: %v", t))
	mx.entries.Store(t, &handler{handler: h, mux: mx})
}

// isHandlerMethod checks if the method is a Executor method.
// A Executor method is a method that has 3 input parameters,
// the first is the receiver, the second is a context.Context,
// and the third is a pointer to a struct that implements the Action or QueryAction interface.
// Example:
//
//	func (uh *UserHandler) Update(ctx context.Context, action *action.UpdateUser) error
func isHandlerMethod(m reflect.Method) bool {
	return m.Type.NumIn() == 3 && isContextType(m.Type.In(1)) && m.Type.NumOut() == 1 && isErrorType(m.Type.Out(0))
}

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

func isContextType(t reflect.Type) bool {
	return t == ctxType
}

func isErrorType(t reflect.Type) bool {
	return t == errType
}

// exec constructs a middleware chain that executes in sequence and only once per context.
func exec(middlewares []middleware, command Middleware) Middleware {
	if len(middlewares) == 0 {
		return command
	}

	return func() Middleware {
		return MiddlewareFunc(func(ctx Context) error {
			idx := ctx.(*BusContext).mwsIdx
			if idx < len(middlewares) {
				ctx.(*BusContext).mwsIdx++
				return middlewares[idx].fn(exec(middlewares, command)).Handle(ctx)
			}
			return command.Handle(ctx)
		})
	}()
}

func chain(op OpType, middlewares []middleware, command Middleware) Middleware {
	// Wrap the end handler with the middleware chain
	mws := filterMiddleware(op, middlewares)

	if len(mws) == 0 {
		return command
	}

	h := mws[len(mws)-1].fn(command)
	for i := len(mws) - 2; i >= 0; i-- {
		h = mws[i].fn(h)
	}

	return h
}

// filterMiddleware returns the middlewares that match the given operation type.
func filterMiddleware(op OpType, middlewares []middleware) []middleware {
	var mws []middleware
	for _, mw := range middlewares {
		if mw.op&op != 0 {
			mws = append(mws, mw)
		}
	}
	return mws
}

type middleware struct {
	op OpType
	fn func(next Middleware) Middleware
}
