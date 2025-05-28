package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a SpacetimeDB client with access to all API endpoints
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	identity   string
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Service interfaces for different API areas
	Identity *IdentityService
	Database *DatabaseService
}

// ClientBuilder provides a builder pattern for constructing clients
type ClientBuilder struct {
	baseURL    string
	token      string
	identity   string
	httpClient *http.Client
	timeout    time.Duration
}

// NewClientBuilder creates a new client builder
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		timeout: 30 * time.Second,
	}
}

// WithBaseURL sets the base URL for the SpacetimeDB instance
func (b *ClientBuilder) WithBaseURL(baseURL string) *ClientBuilder {
	b.baseURL = baseURL
	return b
}

// WithToken sets the authentication token
func (b *ClientBuilder) WithToken(token string) *ClientBuilder {
	b.token = token
	return b
}

// WithIdentity sets the identity
func (b *ClientBuilder) WithIdentity(identity string) *ClientBuilder {
	b.identity = identity
	return b
}

// WithTimeout sets the HTTP client timeout
func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	b.timeout = timeout
	return b
}

// WithHTTPClient sets a custom HTTP client
func (b *ClientBuilder) WithHTTPClient(client *http.Client) *ClientBuilder {
	b.httpClient = client
	return b
}

// Build creates the configured client
func (b *ClientBuilder) Build() (*Client, error) {
	if b.baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	parsedURL, err := url.Parse(b.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	httpClient := b.httpClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: b.timeout,
		}
	}

	client := &Client{
		baseURL:    parsedURL.String(),
		httpClient: httpClient,
		token:      b.token,
		identity:   b.identity,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// Initialize service interfaces
	client.Identity = NewIdentityService(client)
	client.Database = NewDatabaseService(client)

	return client, nil
}

// Close closes the client and its connections
func (c *Client) Close() error {
	c.cancelFunc()
	return nil
}

// Ping tests connectivity to the SpacetimeDB instance
func (c *Client) Ping() error {
	url := fmt.Sprintf("%s/v1/ping", c.baseURL)

	req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error pinging SpacetimeDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetToken returns the current token
func (c *Client) GetToken() string {
	return c.token
}

// SetToken updates the current token
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetIdentity returns the current identity
func (c *Client) GetIdentity() string {
	return c.identity
}

// SetIdentity updates the current identity
func (c *Client) SetIdentity(identity string) {
	c.identity = identity
}

// GetBaseURL returns the base URL
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetHTTPClient returns the underlying HTTP client
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// GetContext returns the client context
func (c *Client) GetContext() context.Context {
	return c.ctx
}

// Helper methods for common HTTP operations

// doRequest performs a basic HTTP request and returns the response
func (c *Client) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return c.httpClient.Do(req)
}

// doAuthenticatedRequest performs an HTTP request with authentication
func (c *Client) doAuthenticatedRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	return c.httpClient.Do(req)
}

// doJSONRequest performs an HTTP request with JSON body and authentication
func (c *Client) doJSONRequest(method, url string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling JSON body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(c.ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// doWASMRequest performs an HTTP request with WASM body and authentication
func (c *Client) doWASMRequest(method, url string, wasmModule []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, method, url, bytes.NewReader(wasmModule))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	req.Header.Set("Content-Type", "application/wasm")

	return c.httpClient.Do(req)
}

// doTextRequest performs an HTTP request with text body and authentication
func (c *Client) doTextRequest(method, url string, text string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, method, url, strings.NewReader(text))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	req.Header.Set("Content-Type", "text/plain")

	return c.httpClient.Do(req)
}

// handleJSONResponse handles a JSON response and unmarshals it
func (c *Client) handleJSONResponse(resp *http.Response, target any) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}

	return nil
}

// handleTextResponse handles a text response and returns the body as string
func (c *Client) handleTextResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

// requiresAuth checks if authentication is required and returns error if not available
func (c *Client) requiresAuth() error {
	if c.token == "" {
		return fmt.Errorf("authentication token is required for this operation")
	}
	return nil
}
