package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type AuthToken struct {
	mu       sync.RWMutex
	token    string
	filePath string
}

const authTokenPrefix = "auth_token="

// AuthTokenOption is a functional option for configuring AuthToken
type AuthTokenOption func(*authTokenConfig)

type authTokenConfig struct {
	configFolder string
	configFile   string
	configRoot   string
}

// WithAuthConfigFolder sets the folder to store the config file in
func WithAuthConfigFolder(folder string) AuthTokenOption {
	return func(c *authTokenConfig) {
		c.configFolder = folder
	}
}

// WithAuthConfigFile sets the name of the config file
func WithAuthConfigFile(file string) AuthTokenOption {
	return func(c *authTokenConfig) {
		c.configFile = file
	}
}

// WithAuthConfigRoot sets the root folder to store the config file in
func WithAuthConfigRoot(root string) AuthTokenOption {
	return func(c *authTokenConfig) {
		c.configRoot = root
	}
}

// NewAuthToken creates and initializes a new AuthToken instance.
// configFolder: The folder to store the config file in. Default is ".spacetime_go_sdk".
// configFile: The name of the config file. Default is "settings.ini".
// configRoot: The root folder to store the config file in. Default is the user's home directory.
func NewAuthToken(options ...AuthTokenOption) (*AuthToken, error) {
	// Set defaults
	config := &authTokenConfig{
		configFolder: ".spacetime_go_sdk",
		configFile:   "settings.ini",
		configRoot:   "",
	}

	// Apply options
	for _, option := range options {
		option(config)
	}

	// Set default config root if not provided
	if config.configRoot == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get home directory: %w", err)
		}
		config.configRoot = homeDir
	}

	// Handle command line client argument
	finalConfigFile := config.configFile
	args := os.Args
	for i, arg := range args {
		if arg == "--client" && i+1 < len(args) {
			parts := strings.Split(config.configFile, ".")
			if len(parts) >= 2 {
				finalConfigFile = fmt.Sprintf("%s_%s.%s", parts[0], args[i+1], parts[1])
			}
			break
		}
	}

	configDir := filepath.Join(config.configRoot, config.configFolder)
	filePath := filepath.Join(configDir, finalConfigFile)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %w", err)
	}

	at := &AuthToken{
		filePath: filePath,
	}

	at.loadToken()
	return at, nil
}

// GetToken returns the auth token that was saved to local storage.
// Returns empty string if never saved.
// When you specify empty string to the SpacetimeDBClient, SpacetimeDB will generate a new identity for you.
func (at *AuthToken) GetToken() string {
	if at == nil {
		return ""
	}
	at.mu.RLock()
	defer at.mu.RUnlock()
	return at.token
}

// SaveToken saves the auth token to local storage.
// SpacetimeDBClient provides this token to you in the onIdentityReceived callback.
func (at *AuthToken) SaveToken(newToken string) error {
	if at == nil {
		return fmt.Errorf("AuthToken not initialized")
	}

	at.mu.Lock()
	defer at.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(at.filePath), 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	var lines []string
	if data, err := os.ReadFile(at.filePath); err == nil {
		lines = strings.Split(string(data), "\n")
	}

	newAuthLine := authTokenPrefix + newToken
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), authTokenPrefix) {
			lines[i] = newAuthLine
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, newAuthLine)
	}

	content := strings.Join(lines, "\n")
	if err := os.WriteFile(at.filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("could not save token: %w", err)
	}

	at.token = newToken
	return nil
}

// GetFilePath returns the path where the auth token is stored (for debugging)
func (at *AuthToken) GetFilePath() string {
	if at == nil {
		return ""
	}
	at.mu.RLock()
	defer at.mu.RUnlock()
	return at.filePath
}

// loadToken loads the token from the file
func (at *AuthToken) loadToken() {
	if at == nil {
		return
	}

	data, err := os.ReadFile(at.filePath)
	if err != nil {
		// File doesn't exist or can't be read, start with empty token
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, authTokenPrefix) {
			at.token = strings.TrimPrefix(line, authTokenPrefix)
			break
		}
	}
}
