package client

import (
	"context"
	"net/http"
)

// Role represents a Warpgate role
type TargetGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// TargetGroupCreateRequest is the request payload for creating a target group
type TargetGroupCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// GetTargetGroup retrieves a specific target group by ID from the Warpgate API.
// Returns nil if the target group is not found.
func (c *Client) GetTargetGroup(ctx context.Context, id string) (*TargetGroup, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/target-groups/"+id, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var targetGroup TargetGroup
	if err := handleResponse(resp, &targetGroup); err != nil {
		return nil, err
	}

	return &targetGroup, nil
}

// CreateTargetGroup creates
func (c *Client) CreateTargetGroup(ctx context.Context, req *TargetGroupCreateRequest) (*TargetGroup, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/target-groups", req)
	if err != nil {
		return nil, err
	}

	var targetGroup TargetGroup
	if err := handleResponse(resp, &targetGroup); err != nil {
		return nil, err
	}

	return &targetGroup, nil
}

// UpdateTargetGroup updates an existing target group's information including name, description, and color.
func (c *Client) UpdateTargetGroup(ctx context.Context, id string, req *TargetGroupCreateRequest) (*TargetGroup, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, "/target-groups/"+id, req)
	if err != nil {
		return nil, err
	}

	var targetGroup TargetGroup
	if err := handleResponse(resp, &targetGroup); err != nil {
		return nil, err
	}

	return &targetGroup, nil
}

// DeleteTargetGroup removes a target group from Warpgate by its ID.
func (c *Client) DeleteTargetGroup(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/target-groups/"+id, nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}
