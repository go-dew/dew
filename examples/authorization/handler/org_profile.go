package handler

import (
	"context"
	"github.com/go-dew/dew/examples/authorization/action"
	"github.com/go-dew/dew/examples/authorization/query"
)

type OrgProfileHandler struct {
}

func NewOrgProfileHandler() *OrgProfileHandler {
	return &OrgProfileHandler{}
}

func (h *OrgProfileHandler) UpdateOrgProfile(_ context.Context, command *action.UpdateOrgProfile) error {
	println("Updating organization profile with name:", command.Name)
	return nil
}

func (h *OrgProfileHandler) GetOrgProfile(_ context.Context, command *query.GetOrgProfile) error {
	command.Result = "Organization Profile"
	return nil
}
