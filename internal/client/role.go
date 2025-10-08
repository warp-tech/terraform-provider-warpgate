// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// Role represents a Warpgate role
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// RoleCreateRequest is the request payload for creating a role
type RoleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GetRoles retrieves all roles from the Warpgate API, optionally filtered by
// the provided search term.
func (c *Client) GetRoles(ctx context.Context, search string) ([]Role, error) {
	path := "/roles"

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Query().Add("search", search)

	resp, err := c.doRequest(ctx, http.MethodGet, req.URL.Path, nil)
	if err != nil {
		return nil, err
	}

	var roles []Role
	if err := handleResponse(resp, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRole retrieves a specific role by ID from the Warpgate API.
// Returns nil if the role is not found.
func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/role/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var role Role
	if err := handleResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// CreateRole creates a new role in Warpgate with the provided name and description.
func (c *Client) CreateRole(ctx context.Context, req *RoleCreateRequest) (*Role, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/roles", req)
	if err != nil {
		return nil, err
	}

	var role Role
	if err := handleResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// UpdateRole updates an existing role's information including name and description.
func (c *Client) UpdateRole(ctx context.Context, id string, req *RoleCreateRequest) (*Role, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/role/%s", id), req)
	if err != nil {
		return nil, err
	}

	var role Role
	if err := handleResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole removes a role from Warpgate by its ID.
func (c *Client) DeleteRole(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/role/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// AddUserRole assigns a role to a user in Warpgate.
func (c *Client) AddUserRole(ctx context.Context, userID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/users/%s/roles/%s", userID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// DeleteUserRole removes a role assignment from a user in Warpgate.
func (c *Client) DeleteUserRole(ctx context.Context, userID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/roles/%s", userID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// GetUserRoles retrieves all roles assigned to a specific user.
func (c *Client) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s/roles", userID), nil)
	if err != nil {
		return nil, err
	}

	var roles []Role
	if err := handleResponse(resp, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// AddTargetRole assigns a role to a target in Warpgate.
func (c *Client) AddTargetRole(ctx context.Context, targetID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/targets/%s/roles/%s", targetID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// DeleteTargetRole removes a role assignment from a target in Warpgate.
func (c *Client) DeleteTargetRole(ctx context.Context, targetID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/targets/%s/roles/%s", targetID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// GetTargetRoles retrieves all roles assigned to a specific target.
func (c *Client) GetTargetRoles(ctx context.Context, targetID string) ([]Role, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/targets/%s/roles", targetID), nil)
	if err != nil {
		return nil, err
	}

	var roles []Role
	if err := handleResponse(resp, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}
