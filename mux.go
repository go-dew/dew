package dew

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type middlewareType int

const (
	mCmd middlewareType = iota
	mDispatch
	mQuery
	mAll
)

// Mux is the main struct where all handlers and middlewares are registered.
type Mux struct {
	parent      *Mux
	inline      bool
	lock        sync.RWMutex
	handler     [ALL]Middleware
	tree        *node
	middlewares [mAll][]middleware
	mHandlers   [mAll]func(ctx Context, fn mHandlerFunc) error

	// context pool
	pool *sync.Pool
}

func (mx *Mux) root() *Mux {
	if mx.parent == nil {
		return mx
	}
	return mx.parent.root()
}

type mHandlerFunc func(ctx Context) error

// newMux returns a newly initialized Mux object that implements the dispatcher interface.
func newMux() *Mux {
	mux := &Mux{tree: &node{}, pool: &sync.Pool{}}
	mux.pool.New = func() interface{} {
		return NewContext()
	}
	return mux
}

// FromContext returns the bus from the context.
func FromContext(ctx context.Context) Bus {
	return ctx.Value(busCtxKey{}).(Bus)
}

type busCtxKey struct{}

// Dispatch executes the action.
// It assumes that all handlers have been registered to the same mux.
func Dispatch(ctx context.Context, actions ...CommandHandler[Action]) error {
	if len(actions) == 0 {
		return nil
	}
	mux := actions[0].Mux().root()

	rctx := mux.pool.Get().(*BusContext)
	rctx.Reset()
	rctx.ctx = context.WithValue(ctx, busCtxKey{}, mux)

	defer mux.pool.Put(rctx)

	return mux.mHandlers[mDispatch](rctx, func(ctx Context) error {
		for _, action := range actions {
			if err := action.Command().(Action).Validate(ctx.Context()); err != nil {
				return fmt.Errorf("%w: %v", ErrValidationFailed, err)
			}
			if err := action.Mux().dispatch(ACTION, ctx, action); err != nil {
				return err
			}
		}
		return nil
	})
}

// Query executes the query and returns the result.
func Query[T QueryAction](ctx context.Context, query CommandHandler[T]) (*T, error) {
	mux := query.Mux().root()

	rctx := mux.pool.Get().(*BusContext)
	rctx.Reset()
	rctx.ctx = context.WithValue(ctx, busCtxKey{}, mux)

	defer mux.pool.Put(rctx)

	if err := mux.mHandlers[mQuery](rctx, func(ctx Context) error {
		return query.Mux().dispatch(QUERY, ctx, query)
	}); err != nil {
		return nil, err
	}

	return query.Command().(*T), nil
}

// QueryAsync executes all queries asynchronously and collects errors.
// It assumes that all handlers have been registered to the same mux.
func QueryAsync(ctx context.Context, queries ...CommandHandler[Command]) error {
	if len(queries) == 0 {
		return nil
	}
	mux := queries[0].Mux().root()

	rctx := mux.pool.Get().(*BusContext) // Get a context from the pool.
	rctx.Reset()
	rctx.ctx = context.WithValue(ctx, busCtxKey{}, mux)

	defer mux.pool.Put(rctx) // Ensure the context is put back into the pool.

	return mux.mHandlers[mQuery](rctx, func(ctx Context) error {
		// Create a goroutine for each query and synchronize with WaitGroup.
		var wg sync.WaitGroup
		errs := make(chan error, len(queries)) // Buffered channel to collect errors from goroutines.

		for _, query := range queries {
			query := query
			wg.Add(1)
			go func(query CommandHandler[Command]) {
				defer wg.Done()
				rctx := mux.pool.Get().(*BusContext) // Get a context from the pool.
				rctx.Reset()
				rctx.Copy(ctx.(*BusContext)) // Copy the context to the new context.

				defer mux.pool.Put(rctx) // Ensure the context is put back into the pool.

				if err := mux.mHandlers[mQuery](rctx, func(ctx Context) error {
					return query.Mux().dispatch(QUERY, ctx, query)
				}); err != nil {
					errs <- err // Send errors to the channel.
				}
			}(query)
		}

		wg.Wait()
		close(errs) // Close the channel after all goroutines are done.

		// Collect errors from the channel.
		var combinedError error
		for err := range errs {
			if combinedError == nil {
				combinedError = err
			} else {
				combinedError = errors.Join(combinedError, err)
			}
		}

		return combinedError
	})

}

// Use appends the middlewares to the mux middleware chain.
// The middleware chain will be executed in the order they were added.
func (mx *Mux) Use(op OpType, middlewares ...func(next Middleware) Middleware) {
	for _, mw := range middlewares {
		mx.middlewares[mCmd] = append(mx.middlewares[mCmd], middleware{op: op, fn: mw})
	}
}

// UseDispatch appends the middlewares to the dispatch middleware chain.
func (mx *Mux) UseDispatch(middlewares ...func(next Middleware) Middleware) {
	mx.addMiddleware(mDispatch, middlewares)
}

// UseQuery appends the middlewares to the query middleware chain.
func (mx *Mux) UseQuery(middlewares ...func(next Middleware) Middleware) {
	mx.addMiddleware(mQuery, middlewares)
}

func (mx *Mux) addMiddleware(m middlewareType, mws []func(next Middleware) Middleware) {
	for _, mw := range mws {
		mx.middlewares[m] = append(mx.middlewares[m], middleware{fn: mw})
	}
}

// Group creates a new mux with a copy of the parent middlewares.
func (mx *Mux) Group(fn func(mx Bus)) Bus {
	child := mx.child()
	if fn != nil {
		fn(child)
	}
	return child
}

// with creates a new mux with the given middlewares.
func (mx *Mux) child() Bus {

	// copy the parent middlewares
	var mws [mAll][]middleware
	for i := range mws {
		mws[i] = make([]middleware, len(mx.middlewares[i]))
		copy(mws[i], mx.middlewares[i])
	}

	return &Mux{
		parent:      mx,
		inline:      true,
		middlewares: mws,
		tree:        mx.tree,
	}
}

type finalHandler interface {
	Handle(ctx Context) error
	Command() Command
}

// dispatch dispatches the command to the appropriate Executor.
func (mx *Mux) dispatch(op OpType, ctx Context, h finalHandler) error {
	hh := mx.handlerFor(op)
	if hh == nil {
		mx.updateRouteHandler(op)
		hh = mx.handlerFor(op)
	}
	ctx.(*BusContext).handler = h
	return hh.Handle(ctx)
}

func (mx *Mux) handlerFor(op OpType) Middleware {
	mx.lock.RLock()
	defer mx.lock.RUnlock()
	return mx.handler[op]
}

func (mx *Mux) newDispatchHandler(m middlewareType, fn func(ctx Context) error) Middleware {
	return exec(mx.middlewares[m], MiddlewareFunc(
		func(ctx Context) error {
			return fn(ctx)
		}))
}

func (mx *Mux) updateRouteHandler(op OpType) {
	mx.lock.Lock()
	defer mx.lock.Unlock()
	mx.handler[op] = chain(op, mx.middlewares[mCmd], MiddlewareFunc(
		func(ctx Context) error {
			return ctx.(*BusContext).handler.Handle(ctx)
		}))
}

func (mx *Mux) updateHandler(m middlewareType) {
	mx.lock.Lock()
	defer mx.lock.Unlock()
	mx.mHandlers[m] = func(ctx Context, fn mHandlerFunc) error {
		return mx.newDispatchHandler(m, func(ctx Context) error {
			return fn(ctx)
		}).Handle(ctx)
	}
}

// Register adds the handler to the mux for the given command type.
func (mx *Mux) Register(handler any) {
	hdlTyp := reflect.TypeOf(handler)
	for i := 0; i < hdlTyp.NumMethod(); i++ {
		mtdTyp := hdlTyp.Method(i)
		if isHandlerMethod(mtdTyp) {
			cmdTyp := mtdTyp.Type.In(2).Elem()
			if cmdTyp.Implements(reflect.TypeOf((*Action)(nil)).Elem()) {
				mx.addHandler(ACTION, cmdTyp, reflect.ValueOf(handler).Method(i).Interface())
			} else if cmdTyp.Implements(reflect.TypeOf((*QueryAction)(nil)).Elem()) {
				mx.addHandler(QUERY, cmdTyp, reflect.ValueOf(handler).Method(i).Interface())
			}
		}
	}
	mx.setupHandler()
}

func (mx *Mux) setupHandler() {
	if mx.mHandlers[mQuery] == nil {
		mx.updateHandler(mQuery)
	}
	if mx.mHandlers[mDispatch] == nil {
		mx.updateHandler(mDispatch)
	}
	if mx.parent != nil {
		mx.parent.setupHandler()
	}
}

func (mx *Mux) addHandler(op OpType, t reflect.Type, h any) {
	mx.tree.insert(op, keyForType(t), &handler{handler: h, mux: mx})
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

// keyForType returns the key for the given type.
func keyForType(typ reflect.Type) string {
	return fmt.Sprintf("%s:%s", typ.PkgPath(), typ.String())
}

// getKey returns the key for the given type.
func getKey[T any]() string {
	var v T
	return keyForType(reflect.TypeOf(v))
}

type middleware struct {
	op OpType
	fn func(next Middleware) Middleware
}
