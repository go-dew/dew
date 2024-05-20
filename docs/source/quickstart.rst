.. _installation

Installation
============

.. contents::
    :local:

To install Dew to your Go project, run the following command in your terminal:

... code-block:: bash

    go get github.com/go-dew/dew

Basic Usage
===========

Here's a simple example to get you started with Dew:

.. code-block:: go

    package main

    import (
        "context"
        "fmt"
        "github.com/go-dew/dew"
    )

    type HelloAction struct {
        Name string
    }

    func (c HelloAction) Validate(_ context.Context) error {
        if c.Name == "" {
            return fmt.Errorf("invalid name")
        }
        return nil
    }

    type HelloHandler struct {}

    func (h *HelloHandler) HandleHelloAction(ctx context.Context, cmd *HelloAction) error {
        println(fmt.Sprintf("Hello, %s!", cmd.Name))
        return nil
    }

    func main() {
        bus := dew.New()
        bus.Register(new(HelloHandler))
        _ = bus.Dispatch(context.Background(), &HelloAction{Name: "Dew"})
    }

This example defines an action `HelloAction` and its handler `HelloHandler`. The action is dispatched to the command bus, which executes the handler and prints a greeting.
