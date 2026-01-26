// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"net/http"
)

// SSHOwnKey represents an SSH host key for the Warpgate server
type SSHOwnKey struct {
	Kind            string `json:"kind"`
	PublicKeyBase64 string `json:"public_key_base64"`
}

// GetSSHOwnKeys retrieves the SSH host keys for the Warpgate server.
// These are the server's own SSH keys that clients use to verify the server identity.
func (c *Client) GetSSHOwnKeys(ctx context.Context) ([]SSHOwnKey, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/ssh/own-keys", nil)
	if err != nil {
		return nil, err
	}

	var keys []SSHOwnKey
	if err := handleResponse(resp, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}
