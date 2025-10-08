package client

import (
	"context"
	"fmt"
	"net/http"
)

// TLSMode represents the TLS mode for a target
type TLSMode string

// TLS mode constants
const (
	// TLSModeDisabled indicates that TLS is disabled
	TLSModeDisabled TLSMode = "Disabled"
	// TLSModePreferred indicates that TLS is preferred but not required
	TLSModePreferred TLSMode = "Preferred"
	// TLSModeRequired indicates that TLS is required
	TLSModeRequired TLSMode = "Required"
)

// TLS represents TLS configuration for a target
type TLS struct {
	Mode   TLSMode `json:"mode"`
	Verify bool    `json:"verify"`
}

// Target represents a Warpgate target
type Target struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	AllowRoles  []string      `json:"allow_roles"`
	Options     TargetOptions `json:"options"`
}

// TargetOptions is a wrapper for the different target option types
type TargetOptions any

// SSHTargetAuth is a wrapper for the different SSH authentication methods
type SSHTargetAuth any

// SSHTargetPasswordAuth represents password authentication for SSH targets
type SSHTargetPasswordAuth struct {
	Kind     string `json:"kind"`
	Password string `json:"password"`
}

// SSHTargetPublicKeyAuth represents public key authentication for SSH targets
type SSHTargetPublicKeyAuth struct {
	Kind string `json:"kind"`
}

// TargetSSHOptions represents options for SSH targets
type TargetSSHOptions struct {
	Kind               string        `json:"kind"`
	Host               string        `json:"host"`
	Port               int           `json:"port"`
	Username           string        `json:"username"`
	AllowInsecureAlgos bool          `json:"allow_insecure_algos,omitempty"`
	Auth               SSHTargetAuth `json:"auth"`
}

// TargetHTTPOptions represents options for HTTP targets
type TargetHTTPOptions struct {
	Kind         string            `json:"kind"`
	URL          string            `json:"url"`
	TLS          TLS               `json:"tls"`
	Headers      map[string]string `json:"headers,omitempty"`
	ExternalHost string            `json:"external_host,omitempty"`
}

// TargetMySQLOptions represents options for MySQL targets
type TargetMySQLOptions struct {
	Kind     string `json:"kind"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	TLS      TLS    `json:"tls"`
}

// TargetPostgresOptions represents options for PostgreSQL targets
type TargetPostgresOptions struct {
	Kind     string `json:"kind"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	TLS      TLS    `json:"tls"`
}

// TargetDataRequest is the request payload for creating/updating a target
type TargetDataRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Options     TargetOptions `json:"options"`
}

// GetTargets retrieves all targets from the Warpgate API, optionally filtered by
// the provided search term.
func (c *Client) GetTargets(ctx context.Context, search string) ([]Target, error) {
	path := "/targets"
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Query().Add("search", search)

	resp, err := c.doRequest(ctx, http.MethodGet, req.URL.Path, nil)
	if err != nil {
		return nil, err
	}

	var targets []Target
	if err := handleResponse(resp, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}

// GetTarget retrieves a specific target by ID from the Warpgate API.
// Returns nil if the target is not found.
func (c *Client) GetTarget(ctx context.Context, id string) (*Target, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/targets/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var target Target
	if err := handleResponse(resp, &target); err != nil {
		return nil, err
	}

	return &target, nil
}

// CreateTarget creates a new target in Warpgate with the provided name, description,
// and configuration options.
func (c *Client) CreateTarget(ctx context.Context, req *TargetDataRequest) (*Target, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/targets", req)
	if err != nil {
		return nil, err
	}

	var target Target
	if err := handleResponse(resp, &target); err != nil {
		return nil, err
	}

	return &target, nil
}

// UpdateTarget updates an existing target's information including name, description,
// and configuration options.
func (c *Client) UpdateTarget(ctx context.Context, id string, req *TargetDataRequest) (*Target, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/targets/%s", id), req)
	if err != nil {
		return nil, err
	}

	var target Target
	if err := handleResponse(resp, &target); err != nil {
		return nil, err
	}

	return &target, nil
}

// DeleteTarget removes a target from Warpgate by its ID.
func (c *Client) DeleteTarget(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/targets/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}
