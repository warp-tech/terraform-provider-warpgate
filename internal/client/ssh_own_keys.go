// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// SSHClientKey represents an SSH key managed by Warpgate (own client keys)
type SSHClientKey struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Kind            string `json:"kind"`
	PublicKey       string `json:"public_key"`
	PublicKeyBase64 string `json:"public_key_base64"`
	IsDefault       bool   `json:"is_default"`
}

// SSHOwnKey is an alias for SSHClientKey for backward compatibility
type SSHOwnKey = SSHClientKey

// ImportSSHClientKeyRequest is the payload for importing an existing SSH client key
type ImportSSHClientKeyRequest struct {
	Label     string `json:"label"`
	SecretKey string `json:"secret_key"`
	IsDefault bool   `json:"is_default"`
}

// GenerateSSHClientKeyRequest is the payload for generating a new SSH client key
type GenerateSSHClientKeyRequest struct {
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

// UpdateSSHClientKeyRequest is the payload for updating an SSH client key
type UpdateSSHClientKeyRequest struct {
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

// GetSSHOwnKeys retrieves all SSH keys for the Warpgate server.
func (c *Client) GetSSHOwnKeys(ctx context.Context) ([]SSHClientKey, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/ssh/own-keys", nil)
	if err != nil {
		return nil, err
	}

	var keys []SSHClientKey
	if err := handleResponse(resp, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// GetSSHOwnKey retrieves a specific SSH key by ID.
func (c *Client) GetSSHOwnKey(ctx context.Context, id string) (*SSHClientKey, error) {
	keys, err := c.GetSSHOwnKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.ID == id {
			return &key, nil
		}
	}

	return nil, nil
}

// ImportSSHOwnKey imports an existing SSH private key into Warpgate.
func (c *Client) ImportSSHOwnKey(ctx context.Context, req *ImportSSHClientKeyRequest) (*SSHClientKey, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/ssh/own-keys", req)
	if err != nil {
		return nil, err
	}

	var key SSHClientKey
	if err := handleResponse(resp, &key); err != nil {
		return nil, err
	}

	return &key, nil
}

// GenerateSSHOwnKey generates a new SSH key pair in Warpgate.
func (c *Client) GenerateSSHOwnKey(ctx context.Context, req *GenerateSSHClientKeyRequest) (*SSHClientKey, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/ssh/own-keys/generate", req)
	if err != nil {
		return nil, err
	}

	var key SSHClientKey
	if err := handleResponse(resp, &key); err != nil {
		return nil, err
	}

	return &key, nil
}

// UpdateSSHOwnKey updates an existing SSH key's label and default status.
func (c *Client) UpdateSSHOwnKey(ctx context.Context, id string, req *UpdateSSHClientKeyRequest) (*SSHClientKey, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/ssh/own-keys/%s", id), req)
	if err != nil {
		return nil, err
	}

	var key SSHClientKey
	if err := handleResponse(resp, &key); err != nil {
		return nil, err
	}

	return &key, nil
}

// DeleteSSHOwnKey deletes an SSH key by ID.
func (c *Client) DeleteSSHOwnKey(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/ssh/own-keys/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}
