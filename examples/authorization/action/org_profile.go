package action

import (
	"context"
	"fmt"
)

var (
	ErrInvalidName = fmt.Errorf("invalid name")
)

type UpdateOrgProfile struct {
	Name string
}

func (c UpdateOrgProfile) Validate(_ context.Context) error {
	if c.Name == "" {
		return ErrInvalidName
	}
	return nil
}

func (c UpdateOrgProfile) Log() string {
	return fmt.Sprintf("Updating organization profile with name: %s", c.Name)
}
