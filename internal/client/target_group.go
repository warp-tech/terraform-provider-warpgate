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
