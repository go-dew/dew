package handler

import (
	"context"
	"fmt"
)

var (
	ErrInvalidName = fmt.Errorf("invalid name")
)

// OrgHandler is a handler for organization commands.
type OrgHandler struct{}

// NewOrgHandler creates a new organization handler.
func NewOrgHandler() *OrgHandler {
	return &OrgHandler{}
}

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

func (h *OrgHandler) UpdateOrg(_ context.Context, command *UpdateOrgAction) error {
	println("Updating organization name:", command.Name)
	return nil
}

// GetOrgDetailsQuery represents the arguments for getting organization details.
type GetOrgDetailsQuery struct{ Result string }

func (h *OrgHandler) GetOrgDetails(_ context.Context, command *GetOrgDetailsQuery) error {
	command.Result = "Get organization details"
	return nil
}
