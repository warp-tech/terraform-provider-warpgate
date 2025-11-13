// Package client provides types and functions for interacting with Warpgate API
package client

import (
	"context"
	"fmt"
	"net/http"
)

// Ticket represents a Warpgate ticket
type Ticket struct {
	ID          string `json:"id"`
	Username    string `json:"username,omitempty"`
	Description string `json:"description,omitempty"`
	Target      string `json:"target,omitempty"`
	UsesLeft    string `json:"uses_left,omitempty"`
	Expiry      string `json:"expiry,omitempty"`
	Created     string `json:"created,omitempty"`
}

// TicketCreateRequest is the request payload for creating a ticket
type TicketCreateRequest struct {
	Username     string `json:"username,omitempty"`
	TargetName   string `json:"target_name,omitempty"`
	Expiry       string `json:"expiry,omitempty"`
	NumberOfUses int    `json:"number_of_uses,omitempty"`
	Description  string `json:"description,omitempty"`
}

// TicketAndSecret represents a ticket along with its secret
type TicketAndSecret struct {
	Ticket Ticket `json:"ticket"`
	Secret string `json:"secret"`
}

// CreateTicket creates a new ticket in Warpgate with the provided parameters.
func (c *Client) CreateTicket(ctx context.Context, req *TicketCreateRequest) (*TicketAndSecret, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/tickets", req)
	if err != nil {
		return nil, err
	}

	var ticketAndSecret TicketAndSecret
	if err := handleResponse(resp, &ticketAndSecret); err != nil {
		return nil, err
	}

	return &ticketAndSecret, nil
}

// DeleteTicket removes a ticket from Warpgate by its ID.
func (c *Client) DeleteTicket(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/tickets/%s", id), nil)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}
