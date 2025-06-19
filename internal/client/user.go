// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// CredentialKind represents a type of credential
type CredentialKind string

// Credential kind constants
const (
	// CredentialKindPassword represents password-based authentication
	CredentialKindPassword CredentialKind = "Password"
	// CredentialKindPublicKey represents public key-based authentication
	CredentialKindPublicKey CredentialKind = "PublicKey"
	// CredentialKindSso represents single sign-on authentication
	CredentialKindTotp CredentialKind = "Totp"
	// CredentialKindSso represents single sign-on authentication
	CredentialKindSso CredentialKind = "Sso"
	// CredentialKindWebUserApproval represents authentication through web user approval
	// This is typically used for interactive confirmation of access
	CredentialKindWebUserApproval CredentialKind = "WebUserApproval"
)

// UserRequireCredentialsPolicy defines the credential policy for a user
type UserRequireCredentialsPolicy struct {
	HTTP     []CredentialKind `json:"http,omitempty"`
	SSH      []CredentialKind `json:"ssh,omitempty"`
	MySQL    []CredentialKind `json:"mysql,omitempty"`
	Postgres []CredentialKind `json:"postgres,omitempty"`
}

// User represents a Warpgate user
type User struct {
	ID               string                        `json:"id"`
	Username         string                        `json:"username"`
	Description      string                        `json:"description,omitempty"`
	CredentialPolicy *UserRequireCredentialsPolicy `json:"credential_policy,omitempty"`
}

// UserCreateRequest is the request payload for creating a user
type UserCreateRequest struct {
	Username    string `json:"username"`
	Description string `json:"description,omitempty"`
}

// UserUpdateRequest is the request payload for updating a user
type UserUpdateRequest struct {
	Username         string                        `json:"username"`
	Description      string                        `json:"description,omitempty"`
	CredentialPolicy *UserRequireCredentialsPolicy `json:"credential_policy,omitempty"`
}

// GetUsers retrieves all users from the Warpgate API, optionally filtered by
// the provided search term.
func (c *Client) GetUsers(ctx context.Context, search string) ([]User, error) {
	path := "/users"
	if search != "" {
		path = fmt.Sprintf("%s?search=%s", path, search)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := handleResponse(resp, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUser retrieves a specific user by ID from the Warpgate API.
// Returns nil if the user is not found.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var user User
	if err := handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user in Warpgate with the provided username and description.
func (c *Client) CreateUser(ctx context.Context, req *UserCreateRequest) (*User, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/users", req)
	if err != nil {
		return nil, err
	}

	var user User
	if err := handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user's information including username, description,
// and credential policy.
func (c *Client) UpdateUser(ctx context.Context, id string, req *UserUpdateRequest) (*User, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/users/%s", id), req)
	if err != nil {
		return nil, err
	}

	var user User
	if err := handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser removes a user from Warpgate by their ID.
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// PasswordCredential represents a password credential for a user
type PasswordCredential struct {
	ID       string `json:"id,omitempty"`
	Password string `json:"password,omitempty"`
}

// AddPasswordCredential adds a password credential to the specified user.
func (c *Client) AddPasswordCredential(ctx context.Context, userID string, password string) (*PasswordCredential, error) {
	req := &PasswordCredential{
		Password: password,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/users/%s/credentials/passwords", userID), req)
	if err != nil {
		return nil, err
	}

	var cred PasswordCredential
	if err := handleResponse(resp, &cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

// DeletePasswordCredential removes a password credential from a user.
func (c *Client) DeletePasswordCredential(ctx context.Context, userID string, credentialID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/credentials/passwords/%s", userID, credentialID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// PublicKeyCredential represents a public key credential for a user
type PublicKeyCredential struct {
	ID               string `json:"id,omitempty"`
	Label            string `json:"label"`
	OpensshPublicKey string `json:"openssh_public_key"`
	DateAdded        string `json:"date_added,omitempty"`
	LastUsed         string `json:"last_used,omitempty"`
}

// AddPublicKeyCredential adds a public key credential to the specified user.
func (c *Client) AddPublicKeyCredential(ctx context.Context, userID string, label, publicKey string) (*PublicKeyCredential, error) {
	req := &PublicKeyCredential{
		Label:            label,
		OpensshPublicKey: publicKey,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/users/%s/credentials/public-keys", userID), req)
	if err != nil {
		return nil, err
	}

	var cred PublicKeyCredential
	if err := handleResponse(resp, &cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

// GetPublicKeyCredentials retrieves all public key credentials for a user.
func (c *Client) GetPublicKeyCredentials(ctx context.Context, userID string) ([]PublicKeyCredential, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s/credentials/public-keys", userID), nil)
	if err != nil {
		return nil, err
	}

	var creds []PublicKeyCredential
	if err := handleResponse(resp, &creds); err != nil {
		return nil, err
	}

	return creds, nil
}

// UpdatePublicKeyCredential updates an existing public key credential.
func (c *Client) UpdatePublicKeyCredential(ctx context.Context, userID string, credentialID string, label string, publicKey string) (*PublicKeyCredential, error) {
	req := &PublicKeyCredential{
		Label:            label,
		OpensshPublicKey: publicKey,
	}

	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/users/%s/credentials/public-keys/%s", userID, credentialID), req)
	if err != nil {
		return nil, err
	}

	var cred PublicKeyCredential
	if err := handleResponse(resp, &cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

// DeletePublicKeyCredential removes a public key credential from a user.
func (c *Client) DeletePublicKeyCredential(ctx context.Context, userID string, credentialID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/credentials/public-keys/%s", userID, credentialID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// SsoCredential represents an SSO credential for a user
type SsoCredential struct {
	ID       string `json:"id,omitempty"`
	Provider string `json:"provider"`
	Email    string `json:"email"`
}

// GetSsoCredentials retrieves all SSO credentials for a user.
func (c *Client) GetSsoCredentials(ctx context.Context, userID string) ([]SsoCredential, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s/credentials/sso", userID), nil)
	if err != nil {
		return nil, err
	}

	var creds []SsoCredential
	if err := handleResponse(resp, &creds); err != nil {
		return nil, err
	}

	return creds, nil
}

// AddSsoCredential adds an SSO credential to the specified user.
func (c *Client) AddSsoCredential(ctx context.Context, userID string, provider, email string) (*SsoCredential, error) {
	req := &SsoCredential{
		Provider: provider,
		Email:    email,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/users/%s/credentials/sso", userID), req)
	if err != nil {
		return nil, err
	}

	var cred SsoCredential
	if err := handleResponse(resp, &cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

// UpdateSsoCredential updates an existing SSO credential.
func (c *Client) UpdateSsoCredential(ctx context.Context, userID string, credentialID string, provider, email string) (*SsoCredential, error) {
	req := &SsoCredential{
		Provider: provider,
		Email:    email,
	}

	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/users/%s/credentials/sso/%s", userID, credentialID), req)
	if err != nil {
		return nil, err
	}

	var cred SsoCredential
	if err := handleResponse(resp, &cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

// DeleteSsoCredential removes an SSO credential from a user.
func (c *Client) DeleteSsoCredential(ctx context.Context, userID string, credentialID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/users/%s/credentials/sso/%s", userID, credentialID), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}
