package client

import (
	"context"
	"fmt"
	"net/http"
)

const (
	SecretBackendAuthToken      = "Token"
	SecretBackendAuthAppRole    = "AppRole"
	SecretBackendAuthKubernetes = "Kubernetes"
)

type SecretBackendAuth struct {
	Method   string `json:"method"`
	Token    string `json:"token,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
	SecretID string `json:"secret_id,omitempty"`
	Role     string `json:"role,omitempty"`
	Mount    string `json:"mount,omitempty"`
}

type SecretBackend struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	BackendType   string            `json:"backend_type"`
	Address       string            `json:"address"`
	Namespace     string            `json:"namespace,omitempty"`
	Auth          SecretBackendAuth `json:"auth"`
	TLSSkipVerify bool              `json:"tls_skip_verify"`
	AllowedPaths  []string          `json:"allowed_paths"`
}

type SecretBackendRequest struct {
	Name          string            `json:"name"`
	BackendType   string            `json:"backend_type"`
	Address       string            `json:"address"`
	Namespace     string            `json:"namespace,omitempty"`
	Auth          SecretBackendAuth `json:"auth"`
	TLSSkipVerify bool              `json:"tls_skip_verify"`
	AllowedPaths  []string          `json:"allowed_paths"`
}

func (c *Client) GetSecretBackends(ctx context.Context) ([]SecretBackend, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/secret-backends", nil)
	if err != nil {
		return nil, err
	}

	var backends []SecretBackend
	if err := handleResponse(resp, &backends); err != nil {
		return nil, err
	}

	return backends, nil
}

func (c *Client) GetSecretBackend(ctx context.Context, id string) (*SecretBackend, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/secret-backends/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, nil
	}

	var backend SecretBackend
	if err := handleResponse(resp, &backend); err != nil {
		return nil, err
	}

	return &backend, nil
}

func (c *Client) CreateSecretBackend(ctx context.Context, req *SecretBackendRequest) (*SecretBackend, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/secret-backends", req)
	if err != nil {
		return nil, err
	}

	var backend SecretBackend
	if err := handleResponse(resp, &backend); err != nil {
		return nil, err
	}

	return &backend, nil
}

func (c *Client) UpdateSecretBackend(ctx context.Context, id string, req *SecretBackendRequest) (*SecretBackend, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/secret-backends/%s", id), req)
	if err != nil {
		return nil, err
	}

	var backend SecretBackend
	if err := handleResponse(resp, &backend); err != nil {
		return nil, err
	}

	return &backend, nil
}

func (c *Client) DeleteSecretBackend(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/secret-backends/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}
