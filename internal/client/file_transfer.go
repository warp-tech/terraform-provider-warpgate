// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// FileTransferPermission represents file transfer permissions for a target-role assignment
type FileTransferPermission struct {
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
