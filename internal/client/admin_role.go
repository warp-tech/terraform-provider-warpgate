// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AdminRolePermissions is the set of admin permissions an admin role grants.
// The field set mirrors the AdminRole object of the Warpgate admin API.
type AdminRolePermissions struct {
	TargetsCreate bool `json:"targets_create"`
	TargetsEdit   bool `json:"targets_edit"`
	TargetsDelete bool `json:"targets_delete"`

	UsersCreate bool `json:"users_create"`
	UsersEdit   bool `json:"users_edit"`
	UsersDelete bool `json:"users_delete"`

	AccessRolesCreate bool `json:"access_roles_create"`
	AccessRolesEdit   bool `json:"access_roles_edit"`
	AccessRolesDelete bool `json:"access_roles_delete"`
	AccessRolesAssign bool `json:"access_roles_assign"`

	SessionsView      bool `json:"sessions_view"`
	SessionsTerminate bool `json:"sessions_terminate"`
	ApproveSessions   bool `json:"approve_sessions"`

	RecordingsView bool `json:"recordings_view"`

	TicketsCreate bool `json:"tickets_create"`
	TicketsDelete bool `json:"tickets_delete"`

	ConfigEdit bool `json:"config_edit"`

	AdminRolesManage bool `json:"admin_roles_manage"`

	TicketRequestsManage bool `json:"ticket_requests_manage"`
}

// AdminRole represents a Warpgate admin role: a named set of admin permissions
// that can be assigned to users. It is distinct from Role, which grants access
// to targets.
type AdminRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AdminRolePermissions
}

// AdminRoleRequest is the request payload for creating or updating an admin role
type AdminRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AdminRolePermissions
}

// GetAdminRoles retrieves all admin roles from the Warpgate API, optionally
// filtered by the provided search term.
func (c *Client) GetAdminRoles(ctx context.Context, search string) ([]AdminRole, error) {
	path := "/admin-roles"
	if search != "" {
		path = fmt.Sprintf("/admin-roles?search=%s", url.QueryEscape(search))
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var roles []AdminRole
	if err := handleResponse(resp, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetAdminRole retrieves a specific admin role by ID from the Warpgate API.
// Returns nil if the admin role is not found.
func (c *Client) GetAdminRole(ctx context.Context, id string) (*AdminRole, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/admin-roles/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, nil
	}

	var role AdminRole
	if err := handleResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// CreateAdminRole creates a new admin role in Warpgate.
func (c *Client) CreateAdminRole(ctx context.Context, req *AdminRoleRequest) (*AdminRole, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin-roles", req)
	if err != nil {
		return nil, err
	}

	var role AdminRole
	if err := handleResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// UpdateAdminRole replaces an existing admin role's name, description and
// permissions.
func (c *Client) UpdateAdminRole(ctx context.Context, id string, req *AdminRoleRequest) (*AdminRole, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/admin-roles/%s", id), req)
	if err != nil {
		return nil, err
	}

	var role AdminRole
	if err := handleResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteAdminRole removes an admin role from Warpgate by its ID. The built-in
// admin role cannot be deleted; the API refuses that with 403.
func (c *Client) DeleteAdminRole(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/admin-roles/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// AddUserAdminRole assigns an admin role to a user in Warpgate.
func (c *Client) AddUserAdminRole(ctx context.Context, userID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/users/%s/admin-roles/%s", userID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// DeleteUserAdminRole removes an admin role assignment from a user in Warpgate.
func (c *Client) DeleteUserAdminRole(ctx context.Context, userID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/admin-roles/%s", userID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// GetUserAdminRoles retrieves all admin roles assigned to a specific user.
// Returns nil if the user is not found.
func (c *Client) GetUserAdminRoles(ctx context.Context, userID string) ([]AdminRole, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s/admin-roles", userID), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, nil
	}

	var roles []AdminRole
	if err := handleResponse(resp, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}
