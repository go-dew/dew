.. _middleware:

Middleware
==========

Middleware in Dew provides a powerful mechanism to intercept and augment the processing of actions and queries within the command bus. Middleware functions can perform a variety of tasks such as logging, validation, transaction management, and more, either before or after the main command handling logic.

Middleware Types
----------------

Dew categorizes middleware into different types based on the operation they are associated with:

- **ACTION**: Middleware that is executed when an action is processed.
- **QUERY**: Middleware that is executed when a query is processed.
- **ALL**: Middleware that is applied to both actions and queries.

Also, middleware can be set for the dispatch of actions and queries:

- **Dispatch Middleware**: Middleware that is executed before the action is dispatched to the handler.
- **Query Middleware**: Middleware that is executed before the query is dispatched to the handler.

Usage
-----

To use middleware in Dew, you attach it to the bus using the ``Use``, ``UseDispatch``, or ``UseQuery`` methods depending on whether you want it to apply to actions, queries, or both.

Adding Middleware to the Bus
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

You can add middleware to the bus at the time of bus creation, before registering any handlers:

.. code-block:: go

    func main() {
        bus := dew.New()

        // Register middleware
        bus.Use(dew.ALL, LoggingMiddleware)
        bus.Use(dew.ACTION, ValidationMiddleware)
        bus.Use(dew.QUERY, AuditMiddleware)

        // Continue setting up bus and handlers
    }

Defining Middleware
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

A middleware in Dew is defined as a function that takes a ``dew.Middleware`` and returns another ``dew.Middleware``. Here's an example of a simple logging middleware:

.. code-block:: go

    func LoggingMiddleware(next dew.Middleware) dew.Middleware {
        return dew.MiddlewareFunc(func(ctx dew.Context) error {
            log.Println("Before executing command or query")
            err := next.Handle(ctx)
            log.Println("After executing command or query")
            return err
        })
    }

Applying Middleware
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

You apply middleware to specific operations using the ``Use``, ``UseDispatch``, and ``UseQuery`` methods of the bus. This design allows for targeted middleware application, enhancing flexibility and control over command and query processing.

Grouping Handlers and Middleware
--------------------------------

Dew allows for grouping of handlers and applying middleware to, allowing granular control just like http routers.

.. code-block:: go

    func main() {
        bus := dew.New()

        // Group middleware and handlers
        bus.Group(func(group dew.Bus) {
            group.Use(dew.ACTION, TransactionMiddleware)
            group.Register(new(UserHandler))
        })
    }
