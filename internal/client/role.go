// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Role represents a Warpgate role
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UserRoleAssignment represents a user's assignment to a role with expiry info
type UserRoleAssignment struct {
	UserID            string  `json:"user_id"`
	RoleID            string  `json:"role_id"`
	RoleName          string  `json:"role_name"`
	GrantedAt         string  `json:"granted_at"`
	GrantedBy         *string `json:"granted_by,omitempty"`
	GrantedByUsername *string `json:"granted_by_username,omitempty"`
	ExpiresAt         *string `json:"expires_at,omitempty"`
	IsExpired         bool    `json:"is_expired"`
	IsActive          bool    `json:"is_active"`
}

// AddUserRoleRequest is the request payload for adding a user role with optional expiry
type AddUserRoleRequest struct {
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// UpdateUserRoleExpiryRequest is the request payload for updating user role expiry
type UpdateUserRoleExpiryRequest struct {
	ExpiresAt *string `json:"expires_at,omitempty"`
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
	if search != "" {
		path = fmt.Sprintf("/roles?search=%s", url.QueryEscape(search))
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
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
		_ = resp.Body.Close()
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

// AddUserRole assigns a role to a user in Warpgate with optional expiry.
func (c *Client) AddUserRole(ctx context.Context, userID, roleID string, expiresAt *string) (*UserRoleAssignment, error) {
	req := &AddUserRoleRequest{
		ExpiresAt: expiresAt,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/users/%s/roles/%s", userID, roleID), req)
	if err != nil {
		return nil, err
	}

	var assignment UserRoleAssignment
	if err := handleResponse(resp, &assignment); err != nil {
		return nil, err
	}

	return &assignment, nil
}

// DeleteUserRole removes a role assignment from a user in Warpgate.
func (c *Client) DeleteUserRole(ctx context.Context, userID, roleID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/roles/%s", userID, roleID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// GetUserRoles retrieves all roles assigned to a specific user with assignment details.
func (c *Client) GetUserRoles(ctx context.Context, userID string) ([]UserRoleAssignment, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s/roles", userID), nil)
	if err != nil {
		return nil, err
	}

	var assignments []UserRoleAssignment
	if err := handleResponse(resp, &assignments); err != nil {
		return nil, err
	}

	return assignments, nil
}

// GetUserRole retrieves a specific user-role assignment.
func (c *Client) GetUserRole(ctx context.Context, userID, roleID string) (*UserRoleAssignment, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s/roles/%s", userID, roleID), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var assignment UserRoleAssignment
	if err := handleResponse(resp, &assignment); err != nil {
		return nil, err
	}

	return &assignment, nil
}

// UpdateUserRoleExpiry updates the expiry time for a user-role assignment.
func (c *Client) UpdateUserRoleExpiry(ctx context.Context, userID, roleID string, expiresAt *string) (*UserRoleAssignment, error) {
	req := &UpdateUserRoleExpiryRequest{
		ExpiresAt: expiresAt,
	}

	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/users/%s/roles/%s/expiry", userID, roleID), req)
	if err != nil {
		return nil, err
	}

	var assignment UserRoleAssignment
	if err := handleResponse(resp, &assignment); err != nil {
		return nil, err
	}

	return &assignment, nil
}

// RemoveUserRoleExpiry removes the expiry time from a user-role assignment.
func (c *Client) RemoveUserRoleExpiry(ctx context.Context, userID, roleID string) (*UserRoleAssignment, error) {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/roles/%s/expiry", userID, roleID), nil)
	if err != nil {
		return nil, err
	}

	var assignment UserRoleAssignment
	if err := handleResponse(resp, &assignment); err != nil {
		return nil, err
	}

	return &assignment, nil
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
