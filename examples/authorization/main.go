package main

import (
	"context"
	"fmt"

	"github.com/go-dew/dew"
	"github.com/go-dew/dew/examples/authorization/handler"
)

var (
	// User IDs for the example.
	AdminID  = 1
	MemberID = 2
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

func main() {
	// Initialize the Command Bus.
	bus := dew.New()

	// Group the handlers and middleware for organization profile authorization.
	bus.Group(func(bus dew.Bus) {
		// Set the authorization middleware
		bus.Use(dew.ACTION, AdminOnly)

		// Register logging middleware.
		bus.Use(dew.ALL, LogCommand)

		// Register the organization profile handler.
		bus.Register(handler.NewOrgHandler())
	})

	// Dispatch an action to update the organization profile. Which should fail because the user is not authorized.
	ctx := ctxWithCurrUser(context.Background(), &CurrentUser{ID: MemberID})
	err := dew.Dispatch(ctx, dew.NewAction(bus, &handler.UpdateOrgAction{Name: "Dew"}))
	println(fmt.Sprintf("Error: %v", err)) // Output: Error: unauthorized

	// Dispatch an action to update the organization profile. Which should succeed because the user is authorized.
	ctx = ctxWithCurrUser(context.Background(), &CurrentUser{ID: AdminID})
	err = dew.Dispatch(ctx, dew.NewAction(bus, &handler.UpdateOrgAction{Name: "Dew"}))
	println(fmt.Sprintf("Error: %v", err)) // Output: Error: <nil>

	// Execute a query to get the organization profile.
	ctx = ctxWithCurrUser(context.Background(), &CurrentUser{ID: MemberID})
	orgProfile, err := dew.Query(ctx, dew.NewQuery(bus, &handler.GetOrgDetailsQuery{}))
	println(
		fmt.Sprintf("Organization Profile: %s, Error: %v", orgProfile, err),
	) // Output: Organization Profile: , Error: <nil>
}
