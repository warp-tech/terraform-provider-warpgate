package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	ID                       string        `json:"id"`
	Name                     string        `json:"name"`
	Description              string        `json:"description,omitempty"`
	GroupId                  string        `json:"group_id,omitempty"`
	RateLimitBytesPerSecond  *int          `json:"rate_limit_bytes_per_second,omitempty"`
	AllowRoles               []string      `json:"allow_roles"`
	Options                  TargetOptions `json:"options"`
	TicketMaxDurationSeconds *int64        `json:"ticket_max_duration_seconds,omitempty"`
	TicketRequestsDisabled   bool          `json:"ticket_requests_disabled"`
	TicketRequireApproval    bool          `json:"ticket_require_approval"`
	TicketMaxUses            *int          `json:"ticket_max_uses,omitempty"`
	// Every write must state the approval gates: Warpgate refuses a target
	// that leaves one out rather than quietly turning it off.
	RequireApproval bool `json:"require_approval"`
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
	Kind  string `json:"kind"`
	KeyID string `json:"key_id,omitempty"`
}

// SSHTargetIamRoleAuth represents IAM role authentication for SSH targets
type SSHTargetIamRoleAuth struct {
	Kind string `json:"kind"`
}

// TargetSSHOptions represents options for SSH targets
type TargetSSHOptions struct {
	Kind               string        `json:"kind"`
	Host               string        `json:"host"`
	Port               int           `json:"port"`
	Username           string        `json:"username"`
	AllowInsecureAlgos bool          `json:"allow_insecure_algos"`
	Auth               SSHTargetAuth `json:"auth"`
	JumpHost           string        `json:"jump_host,omitempty"`
}

// TargetHTTPOptions represents options for HTTP targets
type TargetHTTPOptions struct {
	Kind         string            `json:"kind"`
	URL          string            `json:"url"`
	TLS          TLS               `json:"tls"`
	Headers      map[string]string `json:"headers"`
	ExternalHost string            `json:"external_host,omitempty"`
}

// DatabaseTargetAuth is a wrapper for the MySQL and PostgreSQL authentication methods
type DatabaseTargetAuth any

// DatabaseTargetPasswordAuth represents password authentication for database targets
type DatabaseTargetPasswordAuth struct {
	Kind     string `json:"kind"`
	Password string `json:"password"`
}

// DatabaseTargetIamRoleAuth represents IAM role authentication for database targets
type DatabaseTargetIamRoleAuth struct {
	Kind string `json:"kind"`
}

// TargetMySQLOptions represents options for MySQL targets
type TargetMySQLOptions struct {
	Kind     string             `json:"kind"`
	Host     string             `json:"host"`
	Port     int                `json:"port"`
	Username string             `json:"username"`
	Auth     DatabaseTargetAuth `json:"auth"`
	TLS      TLS                `json:"tls"`
}

// TargetPostgresOptions represents options for PostgreSQL targets
type TargetPostgresOptions struct {
	Kind                string             `json:"kind"`
	Host                string             `json:"host"`
	Port                int                `json:"port"`
	Username            string             `json:"username"`
	DefaultDatabaseName string             `json:"default_database_name,omitempty"`
	IdleTimeout         string             `json:"idle_timeout,omitempty"`
	ProtocolVersion     string             `json:"protocol_version"`
	Auth                DatabaseTargetAuth `json:"auth"`
	TLS                 TLS                `json:"tls"`
}

// KubernetesTargetAuth is a wrapper for the different Kubernetes authentication methods
type KubernetesTargetAuth any

// KubernetesTargetTokenAuth represents token authentication for Kubernetes targets
type KubernetesTargetTokenAuth struct {
	Kind  string `json:"kind"`
	Token string `json:"token"`
}

// KubernetesTargetCertificateAuth represents certificate authentication for Kubernetes targets
type KubernetesTargetCertificateAuth struct {
	Kind        string `json:"kind"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
}

// KubernetesTargetIamRoleAuth represents AWS IAM role authentication for Kubernetes targets
type KubernetesTargetIamRoleAuth struct {
	Kind string `json:"kind"`
}

// TargetVncOptions represents options for VNC targets
type TargetVncOptions struct {
	Kind string        `json:"kind"`
	Host string        `json:"host"`
	Port int           `json:"port"`
	Auth VncTargetAuth `json:"auth"`
}

// VncTargetAuth is a wrapper for the different VNC authentication methods
type VncTargetAuth any

// VncTargetNoneAuth represents a VNC server without authentication
type VncTargetNoneAuth struct {
	Kind string `json:"kind"`
}

// VncTargetPasswordAuth represents password authentication for VNC targets
type VncTargetPasswordAuth struct {
	Kind     string `json:"kind"`
	Password string `json:"password"`
}

// TargetRDPOptions represents options for RDP targets
type TargetRDPOptions struct {
	Kind             string        `json:"kind"`
	Host             string        `json:"host"`
	Port             int           `json:"port"`
	Username         string        `json:"username"`
	Domain           string        `json:"domain,omitempty"`
	Auth             RDPTargetAuth `json:"auth"`
	VerifyTLS        bool          `json:"verify_tls"`
	Compression      string        `json:"compression"`
	InteractiveLogon bool          `json:"interactive_logon"`
	TLSSecurity      string        `json:"tls_security"`
}

// RDPTargetAuth is a wrapper for the different RDP authentication methods
type RDPTargetAuth any

// RDPTargetPasswordAuth represents password authentication for RDP targets
type RDPTargetPasswordAuth struct {
	Kind     string `json:"kind"`
	Password string `json:"password"`
}

// TargetKubernetesOptions represents options for Kubernetes targets
type TargetKubernetesOptions struct {
	Kind       string               `json:"kind"`
	ClusterURL string               `json:"cluster_url"`
	TLS        TLS                  `json:"tls"`
	Auth       KubernetesTargetAuth `json:"auth"`
}

// TargetDataRequest is the request payload for creating/updating a target
type TargetDataRequest struct {
	Name                     string        `json:"name"`
	Description              string        `json:"description,omitempty"`
	GroupId                  string        `json:"group_id,omitempty"`
	RateLimitBytesPerSecond  *int          `json:"rate_limit_bytes_per_second,omitempty"`
	Options                  TargetOptions `json:"options"`
	TicketMaxDurationSeconds *int64        `json:"ticket_max_duration_seconds,omitempty"`
	TicketRequestsDisabled   bool          `json:"ticket_requests_disabled"`
	TicketRequireApproval    bool          `json:"ticket_require_approval"`
	TicketMaxUses            *int          `json:"ticket_max_uses,omitempty"`
	// Every write must state the approval gates: Warpgate refuses a target
	// that leaves one out rather than quietly turning it off.
	RequireApproval bool `json:"require_approval"`
}

// GetTargets retrieves all targets from the Warpgate API, optionally filtered by
// the provided search term.
func (c *Client) GetTargets(ctx context.Context, search string) ([]Target, error) {
	path := "/targets"
	if search != "" {
		path = fmt.Sprintf("/targets?search=%s", url.QueryEscape(search))
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
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
		_ = resp.Body.Close()
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
