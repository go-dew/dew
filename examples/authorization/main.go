package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-dew/dew"
	"github.com/go-dew/dew/examples/authorization/commands/action"
	"github.com/go-dew/dew/examples/authorization/commands/query"
	"github.com/go-dew/dew/examples/authorization/handlers"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

var (
	adminID  = 1
	memberID = 2
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	bus := initializeBus()

	fmt.Println("--- Authorization Example ---")

	if err := runMemberScenario(bus); err != nil {
		return fmt.Errorf("member scenario failed: %w", err)
	}

	if err := runAdminScenario(bus); err != nil {
		return fmt.Errorf("admin scenario failed: %w", err)
	}

	fmt.Println("\n--- Authorization Example finished ---")
	return nil
}

func initializeBus() dew.Bus {
	bus := dew.New()
	bus.Group(func(bus dew.Bus) {
		bus.Use(dew.ACTION, AdminOnly)
		bus.Use(dew.ALL, LogCommand)
		bus.Register(handlers.NewOrgHandler())
	})
	return bus
}

func runMemberScenario(bus dew.Bus) error {
	busContext := dew.NewContext(context.Background(), bus)
	memberContext := authContext(busContext, &CurrentUser{ID: memberID})

	fmt.Println("\n1. Execute a query to get the organization profile (should succeed for member).")
	orgProfile, err := dew.Query(memberContext, &query.GetOrgDetailsQuery{})
	if err != nil {
		return fmt.Errorf("unexpected error in GetOrgDetailsQuery: %w", err)
	}
	fmt.Printf("Organization Profile: %s\n", orgProfile.Result)

	fmt.Println("\n2. Dispatch an action to update the organization profile (should fail for member).")
	_, err = dew.Dispatch(memberContext, &action.UpdateOrgAction{Name: "Foo"})
	if err == nil {
		return fmt.Errorf("expected unauthorized error, got nil")
	}
	if err != ErrUnauthorized {
		return fmt.Errorf("expected unauthorized error, got: %w", err)
	}
	fmt.Printf("Expected unauthorized error: %v\n", err)

	return nil
}

func runAdminScenario(bus dew.Bus) error {
	busContext := dew.NewContext(context.Background(), bus)
	adminContext := authContext(busContext, &CurrentUser{ID: adminID})

	fmt.Println("\n3. Dispatch an action to update the organization profile (should succeed for admin).")
	err := dew.DispatchMulti(adminContext, dew.NewAction(&action.UpdateOrgAction{Name: "Foo"}))
	if err != nil {
		return fmt.Errorf("unexpected error in UpdateOrgAction: %w", err)
	}
	fmt.Println("\nOrganization profile updated successfully.")

	return nil
}
