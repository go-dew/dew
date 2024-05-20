.. _testing:

Testing
=======

With Dew, there's no need to create mocking object for testing.

Example Test Setup
~~~~~~~~~~~~~~~~~~

Here is a basic example showing how to set up a test environment for a Dew application:

.. code-block:: go

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

This example demonstrates how to mock a handler for testing purposes.

Testing Middleware
~~~~~~~~~~~~~~~~~~~~

Testing middleware involves verifying that it behaves as expected both before and after command execution. You can inject middleware into the test bus instance and use assertions to ensure that it performs the correct operations.

.. code-block:: go

    func TestLoggingMiddleware(t *testing.T) {
        loggedMessages := []string{}
        logger := func(message string) {
            loggedMessages = append(loggedMessages, message)
        }

        bus := dew.New()
        bus.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
            return dew.MiddlewareFunc(func(ctx dew.Context) error {
                logger("Before command")
                err := next.Handle(ctx)
                logger("After command")
                return err
            })
        })
        bus.Register(new(MockCommandHandler))

        // Dispatch a test command
        _ = dew.Dispatch(context.Background(), dew.NewAction(bus, &TestCommand{}))

        if len(loggedMessages) != 2 || loggedMessages[0] != "Before command" || loggedMessages[1] != "After command" {
            t.Errorf("Logging middleware did not log correctly")
        }
    }

    type MockCommandHandler struct{}

    func (h *MockCommandHandler) HandleTestCommand(ctx context.Context, cmd *TestCommand) error {
        return nil
    }


