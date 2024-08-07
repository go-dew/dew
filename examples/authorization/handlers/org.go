package handlers

import (
	"context"

	"github.com/go-dew/dew/examples/authorization/commands/action"
	"github.com/go-dew/dew/examples/authorization/commands/query"
)

// OrgHandler is a handler for organization commands.
type OrgHandler struct{}

// NewOrgHandler creates a new organization handler.
func NewOrgHandler() *OrgHandler {
	return &OrgHandler{}
}

func (h *OrgHandler) UpdateOrg(_ context.Context, command *action.UpdateOrgAction) error {
	println("Updating organization name:", command.Name)
	return nil
}

func (h *OrgHandler) GetOrgDetails(_ context.Context, command *query.GetOrgDetailsQuery) error {
	command.Result = "Get organization details"
	return nil
}
