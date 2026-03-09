// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// ParameterValues represents the global parameters retrieved from Warpgate
type ParameterValues struct {
	AllowOwnCredentialManagement    bool `json:"allow_own_credential_management"`
	RateLimitBytesPerSecond         int  `json:"rate_limit_bytes_per_second,omitempty"`
	SSHClientAuthPublickey          bool `json:"ssh_client_auth_publickey"`
	SSHClientAuthPassword           bool `json:"ssh_client_auth_password"`
	SSHClientAuthKeyboardInteractive bool `json:"ssh_client_auth_keyboard_interactive"`
	MinimizePasswordLogin           bool `json:"minimize_password_login"`
}

// ParametersUpdateRequest is the request payload for updating parameters
type ParametersUpdateRequest struct {
	AllowOwnCredentialManagement    bool `json:"allow_own_credential_management"`
	RateLimitBytesPerSecond         int  `json:"rate_limit_bytes_per_second,omitempty"`
	SSHClientAuthPublickey          bool `json:"ssh_client_auth_publickey"`
	SSHClientAuthPassword           bool `json:"ssh_client_auth_password"`
	SSHClientAuthKeyboardInteractive bool `json:"ssh_client_auth_keyboard_interactive"`
	MinimizePasswordLogin           bool `json:"minimize_password_login"`
}

// GetParameters retrieves the global parameters from Warpgate
func (c *Client) GetParameters(ctx context.Context) (*ParameterValues, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/parameters", nil)
	if err != nil {
		return nil, err
	}

	var parameters ParameterValues
	if err := handleResponse(resp, &parameters); err != nil {
		return nil, err
	}

	return &parameters, nil
}

// UpdateParameters updates the global parameters in Warpgate
// Note: The API returns HTTP 201 with no response body, so we fetch the current state after update
func (c *Client) UpdateParameters(ctx context.Context, req *ParametersUpdateRequest) (*ParameterValues, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, "/parameters", req)
	if err != nil {
		return nil, err
	}

	// PUT /parameters returns 201 with no body, so we need to discard the response
	// and fetch the current state instead
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to update parameters: HTTP %d", resp.StatusCode)
	}

	// Fetch the updated parameters
	return c.GetParameters(ctx)
}
