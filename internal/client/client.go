// Package client provides the API client for interacting with the Warpgate API
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
)

// Config contains the configuration for the client
type Config struct {
	Host               string
	Token              string
	Timeout            time.Duration
	InsecureSkipVerify bool
}

// Client is a Warpgate API client
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

// NewClient creates a new Warpgate API client with the provided configuration.
// It validates the host URL and sets up an HTTP client with the specified timeout.
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}

	baseURL, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL: %w", err)
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	return &Client{
		baseURL: baseURL,
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify},
			},
		},
	}, nil
}

// doRequest performs an HTTP request to the Warpgate API with the given method,
// path, and body. It handles URL resolution, request body serialization, and
// authentication via token.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reqURL *url.URL

	if strings.HasPrefix(path, "/") {
		// Create a new URL that has the same scheme and host but combines the paths
		fullPath := c.baseURL.Path
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += strings.TrimPrefix(path, "/")

		reqURL = &url.URL{
			Scheme: c.baseURL.Scheme,
			Host:   c.baseURL.Host,
			Path:   fullPath,
		}
	} else {
		// Path doesn't start with slash, can use normal resolution
		u, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		reqURL = c.baseURL.ResolveReference(u)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("X-Warpgate-Token", c.token)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleResponse processes the API response, checking for errors and unmarshaling
// the response body into the provided result object if applicable.
func handleResponse(resp *http.Response, result any) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API request failed with status %d: (error reading response body: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
