package dew

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrValidationFailed is returned when the command validation fails.
	ErrValidationFailed = fmt.Errorf("validation failed")
)

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
