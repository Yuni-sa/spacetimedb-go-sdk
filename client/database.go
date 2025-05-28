package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// DatabaseService handles all database-related operations
type DatabaseService struct {
	client *Client
}

// NewDatabaseService creates a new database service
func NewDatabaseService(client *Client) *DatabaseService {
	return &DatabaseService{client: client}
}

// DatabaseIdentity represents a SpacetimeDB identity
type DatabaseIdentity struct {
	Identity string `json:"__identity__"`
}

// HostType represents the host type for a database
type HostType struct {
	Wasm []any `json:"Wasm"`
}

// DatabaseInfo represents information about a database
type DatabaseInfo struct {
	DatabaseIdentity DatabaseIdentity `json:"database_identity"`
	OwnerIdentity    DatabaseIdentity `json:"owner_identity"`
	HostType         HostType         `json:"host_type"`
	InitialProgram   string           `json:"initial_program"`
}

// PublishResponse represents the response from publishing a database
type PublishResponse struct {
	Success struct {
		Domain           *string `json:"domain"`
		DatabaseIdentity string  `json:"database_identity"`
		Op               string  `json:"op"` // "created" or "updated"
	} `json:"Success,omitempty"`
	PermissionDenied *struct {
		Name string `json:"name"`
	} `json:"PermissionDenied,omitempty"`
}

// NamesResponse represents the response from getting database names
type NamesResponse struct {
	Names []string `json:"names"`
}

// SetNameResponse represents the response from setting a database name
type SetNameResponse struct {
	Success *struct {
		Domain         string `json:"domain"`
		DatabaseResult string `json:"database_result"`
	} `json:"Success,omitempty"`
	PermissionDenied *struct {
		Domain string `json:"domain"`
	} `json:"PermissionDenied,omitempty"`
}

// SQLResult represents a single SQL query result
type SQLResult struct {
	Schema ProductType `json:"schema"`
	Rows   []any       `json:"rows"`
}

// Publish publishes a new database with no name
func (s *DatabaseService) Publish(wasmModule []byte) (*PublishResponse, error) {
	if err := s.client.requiresAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/database", s.client.baseURL)

	resp, err := s.client.doWASMRequest(http.MethodPost, url, wasmModule)
	if err != nil {
		return nil, err
	}

	var publishResp PublishResponse
	if err := s.client.handleJSONResponse(resp, &publishResp); err != nil {
		return nil, err
	}

	if publishResp.PermissionDenied != nil {
		return &publishResp, fmt.Errorf("permission denied: %s", publishResp.PermissionDenied.Name)
	}

	return &publishResp, nil
}

// PublishTo publishes to a database with the specified name or identity
func (s *DatabaseService) PublishTo(nameOrIdentity string, wasmModule []byte, clear bool) (*PublishResponse, error) {
	if err := s.client.requiresAuth(); err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("%s/v1/database/%s", s.client.baseURL, nameOrIdentity)

	// Add clear parameter if needed
	if clear {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL: %w", err)
		}
		params := url.Values{}
		params.Set("clear", "true")
		parsedURL.RawQuery = params.Encode()
		baseURL = parsedURL.String()
	}

	resp, err := s.client.doWASMRequest(http.MethodPost, baseURL, wasmModule)
	if err != nil {
		return nil, err
	}

	var publishResp PublishResponse
	if err := s.client.handleJSONResponse(resp, &publishResp); err != nil {
		return nil, err
	}

	if publishResp.PermissionDenied != nil {
		return &publishResp, fmt.Errorf("permission denied: %s", publishResp.PermissionDenied.Name)
	}

	return &publishResp, nil
}

// GetInfo retrieves information about a database
func (s *DatabaseService) GetInfo(nameOrIdentity string) (*DatabaseInfo, error) {
	url := fmt.Sprintf("%s/v1/database/%s", s.client.baseURL, nameOrIdentity)

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var dbInfo DatabaseInfo
	if err := s.client.handleJSONResponse(resp, &dbInfo); err != nil {
		return nil, err
	}

	return &dbInfo, nil
}

// Delete deletes a database
func (s *DatabaseService) Delete(nameOrIdentity string) error {
	if err := s.client.requiresAuth(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/database/%s", s.client.baseURL, nameOrIdentity)

	resp, err := s.client.doAuthenticatedRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	return s.client.handleJSONResponse(resp, nil)
}

// GetNames gets the names this database can be identified by
func (s *DatabaseService) GetNames(nameOrIdentity string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/database/%s/names", s.client.baseURL, nameOrIdentity)

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response NamesResponse
	if err := s.client.handleJSONResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Names, nil
}

// AddName adds a new name for this database
func (s *DatabaseService) AddName(nameOrIdentity, newName string) (*SetNameResponse, error) {
	if err := s.client.requiresAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/database/%s/names", s.client.baseURL, nameOrIdentity)

	resp, err := s.client.doTextRequest(http.MethodPost, url, newName)
	if err != nil {
		return nil, err
	}

	var setNameResp SetNameResponse
	if err := s.client.handleJSONResponse(resp, &setNameResp); err != nil {
		return nil, err
	}

	if setNameResp.PermissionDenied != nil {
		return &setNameResp, fmt.Errorf("permission denied: %s", setNameResp.PermissionDenied.Domain)
	}

	return &setNameResp, nil
}

// SetNames sets the list of names for this database
func (s *DatabaseService) SetNames(nameOrIdentity string, names []string) error {
	if err := s.client.requiresAuth(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/database/%s/names", s.client.baseURL, nameOrIdentity)

	resp, err := s.client.doJSONRequest(http.MethodPut, url, names)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 { // http.StatusUnauthorized
		return fmt.Errorf("permission denied")
	}

	return s.client.handleJSONResponse(resp, nil)
}

// GetIdentity gets the identity of a database
func (s *DatabaseService) GetIdentity(nameOrIdentity string) (string, error) {
	url := fmt.Sprintf("%s/v1/database/%s/identity", s.client.baseURL, nameOrIdentity)

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	return s.client.handleTextResponse(resp)
}

// CallReducer invokes a reducer in a database
func (s *DatabaseService) CallReducer(nameOrIdentity, reducerName string, args []any) error {
	if err := s.client.requiresAuth(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/database/%s/call/%s", s.client.baseURL, nameOrIdentity, reducerName)

	resp, err := s.client.doJSONRequest(http.MethodPost, url, args)
	if err != nil {
		return err
	}

	return s.client.handleJSONResponse(resp, nil)
}

// GetSchema gets a schema for a database
func (s *DatabaseService) GetSchema(nameOrIdentity string, _ *int) (RawModuleDef, error) {
	baseURL := fmt.Sprintf("%s/v1/database/%s/schema", s.client.baseURL, nameOrIdentity)

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return RawModuleDef{}, fmt.Errorf("error parsing URL: %w", err)
	}
	params := url.Values{}
	params.Set("version", "9")
	parsedURL.RawQuery = params.Encode()

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return RawModuleDef{}, err
	}

	var schema RawModuleDef
	if err := s.client.handleJSONResponse(resp, &schema); err != nil {
		return RawModuleDef{}, err
	}

	return schema, nil
}

// GetLogs retrieves logs from a database
func (s *DatabaseService) GetLogs(nameOrIdentity string, numLines *int, follow bool) (string, error) {
	if err := s.client.requiresAuth(); err != nil {
		return "", err
	}

	baseURL := fmt.Sprintf("%s/v1/database/%s/logs", s.client.baseURL, nameOrIdentity)

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	params := url.Values{}
	if numLines != nil {
		params.Set("num_lines", fmt.Sprintf("%d", *numLines))
	}
	if follow {
		params.Set("follow", "true")
	}
	parsedURL.RawQuery = params.Encode()

	resp, err := s.client.doAuthenticatedRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	return s.client.handleTextResponse(resp)
}

// ExecuteSQL runs a SQL query against a database
func (s *DatabaseService) ExecuteSQL(nameOrIdentity string, queries []string) ([]SQLResult, error) {
	if err := s.client.requiresAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/database/%s/sql", s.client.baseURL, nameOrIdentity)

	// Join queries with semicolon
	sqlString := strings.Join(queries, ";")

	resp, err := s.client.doTextRequest(http.MethodPost, url, sqlString)
	if err != nil {
		return nil, err
	}

	var results []SQLResult
	if err := s.client.handleJSONResponse(resp, &results); err != nil {
		return nil, err
	}

	return results, nil
}
