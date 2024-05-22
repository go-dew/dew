# Dew: A Lightweight, Pragmatic Command Bus Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/go-dew/dew.svg)](https://pkg.go.dev/github.com/go-dew/dew)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-dew/dew)](https://goreportcard.com/report/github.com/go-dew/dew)
[![codecov](https://codecov.io/gh/go-dew/dew/branch/main/graph/badge.svg?token=3ZQZQZQZQZ)](https://codecov.io/gh/go-dew/dew)

<img src="assets/dew.png" alt="dew logo" style="width: 200px;" />

Dew is a command bus library for Go, designed to enhance developer experience and productivity. It utilizes the [command-oriented interface](https://martinfowler.com/bliki/CommandOrientedInterface.html) pattern, which allows for separation of concerns, modularization, and better readability of the codebase, eliminating unnecessary cognitive load.

## Features

- **Lightweight**: Clocks around 600 LOC with minimalistic design.
- **Pragmatic and Ergonomic**: Focused on developer experience and productivity.
- **Production Ready**: 100% test coverage.
- **Zero Dependencies**: No external dependencies.
- **Fast**: See [benchmarks](#benchmarks).

## Installation

```bash
go get github.com/go-dew/dew
```

## Example

See [examples](examples) for more detailed examples.

It's as easy as:

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-dew/dew"
)

// HelloAction is a simple action that greets the user.
type HelloAction struct {
    Name string
}

// Validate checks if the name is valid.
func (c HelloAction) Validate(_ context.Context) error {
    if c.Name == "" {
        return fmt.Errorf("invalid name")
    }
    return nil
}

func main() {
    // Initialize the Command Bus.
    bus := dew.New()

    // Register the handler for the HelloAction.
    bus.Register(new(HelloHandler))
    
    // Alternatively, you can use the HandlerFunc to register the handler.
    // bus.Register(dew.HandlerFunc[HelloAction](func(ctx context.Context, cmd *HelloAction) error {
    //     println(fmt.Sprintf("Hello, %s!", cmd.Name)) // Output: Hello, Dew!
    //     return nil
    // }))

    // Dispatch the action.
    _ = dew.Dispatch(context.Background(), dew.NewAction(bus, &HelloAction{Name: "Dew"}))
}

type HelloHandler struct {}
func (h *HelloHandler) HandleHelloAction(ctx context.Context, cmd *HelloAction) error {
    println(fmt.Sprintf("Hello, %s!", cmd.Name)) // Output: Hello, Dew!
    return nil
}
```

## Terminology

Dew uses the following terminology:

- **Action**: Operations that change the application state. We use the term "Action" to avoid confusion with similar terms in Go. It's equivalent to what is commonly known as a "Command" in [Command Query Separation (CQS)](https://en.wikipedia.org/wiki/Command%E2%80%93query_separation) and [Command Query Responsibility Segregation (CQRS)](https://martinfowler.com/bliki/CQRS.html) patterns.
- **Query**: Operations that retrieve data.
- **Middleware**: Functions that execute logic (e.g., logging, authorization, transaction management) before and after command execution.
- **Bus**: Manages registration of handlers and routing of actions and queries to their respective handlers.

## What is Command Bus?

A command bus is a design pattern that separates the execution of commands from their processing logic. It decouples the sender of a command from the handler, enhancing code modularization and separation of concerns.

You can find more about the command bus pattern in the following articles:

- [Command Oriented Interface by Martin Fowler](https://martinfowler.com/bliki/CommandOrientedInterface.html)
- [What is a command bus and why should you use it?](https://barryvanveen.nl/articles/49-what-is-a-command-bus-and-why-should-you-use-it)
- [Laravel Command Bus Pattern](https://laravel.com/docs/5.0/bus)

## Motivation

I've been working on multiple complex backend applications built in Go over the years, and looking for a way to make the code more readable, maintainable, and more fun to work with. I believe Command Bus architecture could be an answer to this problem. However, I couldn't find a library that fits my needs, so I decided to create Dew.

There are several benefits to using Dew:

- It provides a bus interface that utilizes Go's generics to handle commands and queries, allowing for better performance and ease of use.
- The middleware system allows for adding features like logging, authorization, and transaction management for groups of handlers with granular control. See [middleware example](#middleware) and [authorization example](examples/authorization/main.go) for more details.
- The unified bus interface eliminates the need for creating and managing a clutter of mock objects of different interfaces, making the unit tests more readable and fun to work with. See [testing example](#testing-example-mocking-command-handlers) for more details.
- With its built-in support for asynchronous queries, Dew can handle multiple queries concurrently, reducing the time to retrieve data from multiple sources. See [QueryAsync example](#executing-queries) for more details.

Dew is designed to be lightweight with zero dependencies, making it easy to integrate into any Go project.

## A Convention for Actions and Queries

Dew relies on a convention for `Action` and `Query` interfaces:

- **Action Interface**: Each action in Dew must implement a `Validate` method, as defined by the `Action` interface. This `Validate` method is responsible for checking that the action's data is correct before it is processed.
- **Query Interface**: Each query in any struct that implements the `Query` interface, which is an empty interface. Queries do not need a `Validate` method because they do not change the state of the application.

Here's a simple example of how both interfaces are defined and used:

```go
type MyAction struct {
    Amount int
}

// Validate implements the Action interface
func (a *MyAction) Validate(ctx context.Context) error {
    if a.Amount <= 0 {
        return fmt.Errorf("amount must be greater than zero")
    }
    return nil
}

type MyQuery struct {
    AccountID string
}

// MyQuery does not need a Validate method because it does not change state
```

Also, we use the function `dew.Dispatch` to send actions and `dew.Query` to send queries to the bus. The bus will then route the action or query to the appropriate handler based on the action or query type. The reason for using different functions for actions and queries is to make the code more readable and simpler to work with. You will see this when you start using Dew in your projects.

## Usage

### Setting Up the Bus

Create a bus and register handlers:

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-dew/dew"
)

func main() {
    bus := dew.New()
    
    // Register handlers
    bus.Register(new(MyCommandHandler))
}

type MyCommandHandler struct {}

func (h *MyCommandHandler) HandleMyCommand(ctx context.Context, cmd *MyCommand) error {
    // handle command
    fmt.Println("Handling command:", cmd)
    return nil
}

type MyCommand struct {
    Message string
}
```

### Dispatching Commands

Use the `Dispatch` function to send commands:

```go
func main() {
    ctx := context.Background()
    bus := dew.New()
    bus.Register(new(MyCommandHandler))

    cmd := &MyCommand{Message: "Hello, Dew!"}
    if err := dew.Dispatch(ctx, dew.NewAction(cmd)); err != nil {
        fmt.Println("Error dispatching command:", err)
    }
}
```

### Executing Queries

`Query` handling example:

```go

type MyQuery struct {
    Question string
    Result string
}

type MyQueryHandler struct {}

func (h *MyQueryHandler) HandleMyQuery(ctx context.Context, query *MyQuery) error {
    // Return query result
    query.Result = "Dew is a command bus library for Go."
    return nil
}

func main() {
    ctx := context.Background()
    bus := dew.New()
    bus.Register(new(MyQueryHandler))

    result, err := dew.Query(ctx, dew.NewQuery(&MyQuery{Question: "What is Dew?"}))
    if err != nil {
        fmt.Println("Error executing query:", err)
    } else {
        fmt.Println("Query result:", result.Result)
    }
}
```

Dew provides `QueryAsync`, which allows for handling multiple queries concurrently.

`QueryAsync` usage example:

```go
type AccountQuery struct {
    AccountID string
    Result    float64
}

type WeatherQuery struct {
    City   string
    Result string
}

type AccountQueryHandler struct {}
type WeatherQueryHandler struct {}

func (h *AccountQueryHandler) HandleAccountQuery(ctx context.Context, query *AccountQuery) error {
    // Logic to retrieve account balance
    query.Result = 10234.56 // Simulated balance
    return nil
}

func (h *WeatherQueryHandler) HandleWeatherQuery(ctx context.Context, query *WeatherQuery) error {
    // Logic to fetch weather forecast
    query.Result = "Sunny with a chance of rain" // Simulated forecast
    return nil
}

func main() {
    ctx := context.Background()
    bus := dew.New()
    bus.Register(new(AccountQueryHandler))
    bus.Register(new(WeatherQueryHandler))

    accountQuery := &AccountQuery{AccountID: "12345"}
    weatherQuery := &WeatherQuery{City: "New York"}

    if err := dew.QueryAsync(ctx, dew.NewQuery(accountQuery), dew.NewQuery(weatherQuery)); err != nil {
        fmt.Println("Error executing queries:", err)
    } else {
        fmt.Println("Account Balance for ID 12345:", accountQuery.Result)
        fmt.Println("Weather in New York:", weatherQuery.Result)
    }
}
```

### Middleware

Middleware can be used to execute logic before and after command or query execution. Here is an example of a simple logging middleware:

```go
func loggingMiddleware(next dew.Middleware) dew.Middleware {
    return dew.MiddlewareFunc(func(ctx dew.Context) error {
        fmt.Println("Before executing command")
        err := next.Handle(ctx)
        fmt.Println("After executing command")
        return err
    })
}

func main() {
    bus := dew.New()
    bus.Use(dew.ACTION, loggingMiddleware)
    bus.Register(new(MyCommandHandler))
}
```

Here is the interface for middleware:

```go
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
```

### Grouping Handlers and Applying Middleware

It's easy to group handlers and apply middleware to a group. You can also nest groups to apply middleware to a subset of handlers. It allows for a clean separation of concerns and reduces code duplication across handlers.

Here is an example of grouping handlers and applying middleware:

```go
func main() {
    bus := dew.New()
    bus.Group(func(bus dew.Bus) {
        // Transaction middleware
        bus.Use(dew.ACTION, middleware.Transaction)
        // Logger middleware
        bus.Use(dew.ALL, middleware.Logger)
        // Register handlers
        bus.Register(new(UserProfileHandler))
        
        // Sub-grouping
        bus.Group(func(g dew.Bus) {
            // Tracing middleware
            bus.Use(dew.ACTION, middleware.Tracing)
            // Register sensitive handlers
            bus.Register(new(SensitiveCommandHandler))
        })
        
        // Register more handlers
    })
}
```

### Notes about Middleware

- Middleware for handlers can be applied per command or query, based on the `dew.ACTION`, `dew.QUERY` and `dew.ALL` constants.
- Middleware can be applied multiple times because they are executed per command or query. So make sure the middleware is idempotent when necessary.
- Middleware for `Dispatch` and `Query` functions can be configured using the `UseDispatch()` and `UseQuery()` methods on the bus. This middleware is executed once per `Dispatch` or `Query` call.

## Middleware Examples: Handling Transactions in Dispatch

Here is an example of a middleware that starts a transaction at the beginning of a command dispatch and rolls it back if any error occurs during the command's execution.

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-dew/dew"
    "database/sql"
)

// TransactionalMiddleware creates a middleware for handling transactions
func TransactionalMiddleware(db *sql.DB) func(next dew.Middleware) dew.Middleware {
    return func(next dew.Middleware) dew.Middleware {
        return dew.MiddlewareFunc(func(ctx dew.Context) error {
            // Check if a transaction is already present in the context
            if tx, ok := ctx.Context().Value("tx").(*sql.Tx); ok && tx != nil {
                // Transaction already exists, proceed without creating a new one
                return next.Handle(ctx)
            }

            // Start a new transaction
            tx, err := db.BeginTx(ctx.Context(), nil)
            if err != nil {
                return fmt.Errorf("failed to begin transaction: %w", err)
            }

            // Attach the transaction to the context
            txCtx := context.WithValue(ctx.Context(), "tx", tx)
            ctx = ctx.WithContext(txCtx)

            // Execute the command
            err = next.Handle(ctx)
            if err != nil {
                // Roll back the transaction in case of an error
                if rbErr := tx.Rollback(); rbErr != nil {
                    return fmt.Errorf("rollback failed: %w", rbErr)
                }
                return err
            }

            // Commit the transaction if everything went well
            if commitErr := tx.Commit(); commitErr != nil {
                return fmt.Errorf("commit failed: %w", commitErr)
            }

            return nil
        })
    }
}

func main() {
    db, err := sql.Open("driver-name", "database-url")
    if err != nil {
        panic("failed to connect database")
    }
    defer db.Close()

    bus := dew.New()
    bus.UseDispatch(TransactionalMiddleware(db))

    // Register your handlers and continue with application setup
}
```

## Testing Example: Mocking Command Handlers

To mock command handlers for testing, you can create a new bus instance and register the mock handlers.

```go
package example_test

import (
    "context"
    "github.com/go-dew/dew"
    "github.com/your/application/internal/action"
    "testing"
)

func TestExample(t *testing.T) {
    // Create a new bus instance
    mux := dew.New()
    
    // Register your mock handlers
    mux.Register(dew.HandlerFunc[action.CreateUser](
        func(ctx context.Context, command *action.CreateUser) error {
            // mock logic
            return nil
        },
    ))
    
    // test your code
}
```

## Benchmarks

Results as of May 23, 2024 with Go 1.22.2 on darwin/arm64

```shell
BenchmarkMux/query-12            3012015               393.5 ns/op           168 B/op          7 allocs/op
BenchmarkMux/dispatch-12         2854291               419.1 ns/op           192 B/op          8 allocs/op
BenchmarkMux/query-with-middleware-12            2981778               407.8 ns/op           168 B/op          7 allocs/op
BenchmarkMux/dispatch-with-middleware-12         2699398               446.8 ns/op           192 B/op          8 allocs/op
```

## Contributing

We welcome contributions to Dew! Please see the [contribution guide](CONTRIBUTING.md) for more information.

## Credits

- The implementation of Trie data structure is inspired by [go-chi/chi](https://github.com/go-chi/chi).

## License

Licensed under [MIT License](https://github.com/go-dew/dew/blob/main/LICENSE)
