.. _dew

Dew
====

.. contents::
    :local:

Dew is a command oriented interface library + Middleware System for Go. With the unified interface, we can be free from the cognitive load for maintaining different interfaces for each operation handler or domain logic.

.. _minimal-example:

Minimum Example
===============

It's easy to get started. Here's a simple example that demonstrates how to use Dew to dispatch an action.

.. code-block:: go

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


See the `example <https://github.com/go-dew/dew/blob/main/examples/authorization/main.go>`_ for a more practical example.
