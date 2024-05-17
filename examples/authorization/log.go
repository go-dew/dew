package main

import "github.com/go-dew/dew"

// Logger is an interface to log the command.
type Logger interface {
	// Log returns a string representation of the command.
	Log() string
}

func LogCommand(next dew.Middleware) dew.Middleware {
	return dew.MiddlewareFunc(func(ctx dew.Context) error {
		if cmd, ok := ctx.Command().(Logger); ok {
			println(cmd.Log()) // Output: Updating organization profile with name: Dew
		}
		// Continue to the next middleware or handler.
		return next.Handle(ctx)
	})
}
