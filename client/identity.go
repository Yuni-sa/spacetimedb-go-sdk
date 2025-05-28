package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// IdentityService handles all identity-related operations
type IdentityService struct {
	client *Client
}

// NewIdentityService creates a new identity service
func NewIdentityService(client *Client) *IdentityService {
	return &IdentityService{client: client}
}

// IdentityResponse represents the response from identity creation
type IdentityResponse struct {
	Identity string `json:"identity"`
	Token    string `json:"token"`
}

// WebSocketTokenResponse represents the response from websocket token generation
type WebSocketTokenResponse struct {
	Token string `json:"token"`
}

// DatabasesResponse represents the response from listing databases
type DatabasesResponse struct {
	Addresses []string `json:"addresses"`
}

// Create creates a new identity and returns the identity and token
func (s *IdentityService) Create() (*IdentityResponse, error) {
	url := fmt.Sprintf("%s/v1/identity", s.client.baseURL)

	resp, err := s.client.doRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	var identityResp IdentityResponse
	if err := s.client.handleJSONResponse(resp, &identityResp); err != nil {
		return nil, err
	}

	return &identityResp, nil
}

// CreateWebSocketToken generates a short-lived access token for use in untrusted contexts
func (s *IdentityService) CreateWebSocketToken() (*WebSocketTokenResponse, error) {
	if err := s.client.requiresAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/identity/websocket-token", s.client.baseURL)

	resp, err := s.client.doAuthenticatedRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	var tokenResp WebSocketTokenResponse
	if err := s.client.handleJSONResponse(resp, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// GetPublicKey fetches the public key used by the database to verify tokens
func (s *IdentityService) GetPublicKey() (string, error) {
	url := fmt.Sprintf("%s/v1/identity/public-key", s.client.baseURL)

	resp, err := s.client.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	return s.client.handleTextResponse(resp)
}

// SetEmail associates an email with a Spacetime identity
func (s *IdentityService) SetEmail(identity, email string) error {
	if err := s.client.requiresAuth(); err != nil {
		return err
	}

	baseURL := fmt.Sprintf("%s/v1/identity/%s/set-email", s.client.baseURL, identity)

	// Add email as query parameter
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}

	params := url.Values{}
	params.Set("email", email)
	parsedURL.RawQuery = params.Encode()

	resp, err := s.client.doAuthenticatedRequest(http.MethodPost, parsedURL.String(), nil)
	if err != nil {
		return err
	}

	return s.client.handleJSONResponse(resp, nil)
}

// Verify verifies an identity and token pair
func (s *IdentityService) Verify(identity string) error {
	if err := s.client.requiresAuth(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/identity/%s/verify", s.client.baseURL, identity)

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Special handling for verify endpoint status codes
	switch resp.StatusCode {
	case http.StatusNoContent: // http.StatusNoContent
		return nil // Valid token and identity match
	case http.StatusBadRequest: // http.StatusBadRequest
		return fmt.Errorf("token is valid but does not match the identity")
	case http.StatusUnauthorized: // http.StatusUnauthorized
		return fmt.Errorf("token is invalid or no authorization header provided")
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// GetDatabases returns a list of databases owned by an identity
func (s *IdentityService) GetDatabases(identity string) ([]string, error) {
	if err := s.client.requiresAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/identity/%s/databases", s.client.baseURL, identity)

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response DatabasesResponse
	if err := s.client.handleJSONResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Addresses, nil
}
