package action

import (
	"context"
	"fmt"
)

var (
	ErrInvalidName = fmt.Errorf("invalid name")
)

// UpdateOrgAction represents the arguments for updating an organization.
type UpdateOrgAction struct{ Name string }

func (c UpdateOrgAction) Validate(_ context.Context) error {
	if c.Name == "" {
		return ErrInvalidName
	}
	return nil
}

func (c UpdateOrgAction) Log() string {
	return fmt.Sprintf("Updating organization with name: %s", c.Name)
}
