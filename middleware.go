package dew

// Middleware is an interface for handling middleware.
type Middleware interface {
	// Handle executes the middleware.
	Handle(ctx Context) error
}

// MiddlewareFunc is a type adapter to convert a function to a Middleware.
type MiddlewareFunc func(ctx Context) error

// Handle calls the function h(ctx, command).
func (h MiddlewareFunc) Handle(ctx Context) error {
	return h(ctx)
}
