// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// FileTransferPermission represents file transfer permissions for a target-role assignment.
// With the new inheritance model, AllowFileUpload and AllowFileDownload are nullable:
// - nil = inherit from role defaults
// - true/false = explicit override
type FileTransferPermission struct {
	AllowFileUpload   *bool    `json:"allow_file_upload"`
	AllowFileDownload *bool    `json:"allow_file_download"`
	AllowedPaths      []string `json:"allowed_paths,omitempty"`
	BlockedExtensions []string `json:"blocked_extensions,omitempty"`
	MaxFileSize       *int64   `json:"max_file_size,omitempty"`
}

// RoleFileTransferDefaults represents file transfer permission defaults for a role.
// These are the default values used when target-role permissions are set to inherit (nil).
type RoleFileTransferDefaults struct {
	AllowFileUpload   bool     `json:"allow_file_upload"`
	AllowFileDownload bool     `json:"allow_file_download"`
	AllowedPaths      []string `json:"allowed_paths,omitempty"`
	BlockedExtensions []string `json:"blocked_extensions,omitempty"`
	MaxFileSize       *int64   `json:"max_file_size,omitempty"`
}

// GetTargetRoleFileTransferPermission retrieves file transfer permissions for a target-role assignment.
func (c *Client) GetTargetRoleFileTransferPermission(ctx context.Context, targetID, roleID string) (*FileTransferPermission, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/targets/%s/roles/%s/file-transfer", targetID, roleID), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var perm FileTransferPermission
	if err := handleResponse(resp, &perm); err != nil {
		return nil, err
	}

	return &perm, nil
}

// UpdateTargetRoleFileTransferPermission updates file transfer permissions for a target-role assignment.
func (c *Client) UpdateTargetRoleFileTransferPermission(ctx context.Context, targetID, roleID string, req *FileTransferPermission) (*FileTransferPermission, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/targets/%s/roles/%s/file-transfer", targetID, roleID), req)
	if err != nil {
		return nil, err
	}

	var perm FileTransferPermission
	if err := handleResponse(resp, &perm); err != nil {
		return nil, err
	}

	return &perm, nil
}

// GetRoleFileTransferDefaults retrieves file transfer default permissions for a role.
func (c *Client) GetRoleFileTransferDefaults(ctx context.Context, roleID string) (*RoleFileTransferDefaults, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/role/%s/file-transfer", roleID), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var defaults RoleFileTransferDefaults
	if err := handleResponse(resp, &defaults); err != nil {
		return nil, err
	}

	return &defaults, nil
}

// UpdateRoleFileTransferDefaults updates file transfer default permissions for a role.
func (c *Client) UpdateRoleFileTransferDefaults(ctx context.Context, roleID string, req *RoleFileTransferDefaults) (*RoleFileTransferDefaults, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/role/%s/file-transfer", roleID), req)
	if err != nil {
		return nil, err
	}

	var defaults RoleFileTransferDefaults
	if err := handleResponse(resp, &defaults); err != nil {
		return nil, err
	}

	return &defaults, nil
}
