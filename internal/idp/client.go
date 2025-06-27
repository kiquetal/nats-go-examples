// Package idp provides functionality for interacting with identity providers
package idp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// TokenResponse represents a response from the IDP with token information
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// Client represents an IDP client for obtaining tokens
type Client struct {
	baseURL       string
	tokenEndpoint string
	httpClient    *http.Client
}

// ClientCredentials holds the credentials for a client
type ClientCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"` // Added scope field
}

// ClientOption represents a function that modifies a Client
type ClientOption func(*Client)

// WithTokenEndpoint sets a custom token endpoint path
func WithTokenEndpoint(path string) ClientOption {
	return func(c *Client) {
		c.tokenEndpoint = path
	}
}

// WithTimeout sets a custom HTTP timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// Configuration constants
const (
	DefaultBaseURL       = "https://idp.example.com"
	DefaultTokenEndpoint = "/realms/phoenix/protocol/openid-connect/token"
)

// NewClient creates a new IDP client with the provided options
func NewClient(baseURL string, options ...ClientOption) *Client {
	// Check for environment variable overrides
	if envURL := os.Getenv("IDP_URL"); envURL != "" {
		baseURL = envURL
	}

	tokenEndpoint := DefaultTokenEndpoint
	if envTokenPath := os.Getenv("IDP_TOKEN_PATH"); envTokenPath != "" {
		tokenEndpoint = envTokenPath
	}

	client := &Client{
		baseURL:       baseURL,
		tokenEndpoint: tokenEndpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// GetTokenWithClientCredentials obtains a token using client credentials
func (c *Client) GetTokenWithClientCredentials(credentials *ClientCredentials) (*TokenResponse, error) {
	// Create form data
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", credentials.ClientID)
	formData.Set("client_secret", credentials.ClientSecret)

	// Add scope if provided
	if credentials.Scope != "" {
		formData.Set("scope", credentials.Scope)
	}

	// Create full token endpoint URL
	tokenURL := fmt.Sprintf("%s%s", c.baseURL, c.tokenEndpoint)

	// Create request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IDP returned error status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// SimulateTokenRetrieval is a mock function that simulates retrieving a token
// This is useful for testing without an actual IDP
func (c *Client) SimulateTokenRetrieval(credentials *ClientCredentials) (*TokenResponse, error) {
	// For simulation purposes, create a fake token based on the client ID
	fakeToken := fmt.Sprintf("fake-token-%s-%d", credentials.ClientID, time.Now().Unix())

	// Simulate network delay
	time.Sleep(200 * time.Millisecond)

	// Include scope in response if provided
	var scope string
	if credentials.Scope != "" {
		scope = credentials.Scope
	}

	return &TokenResponse{
		AccessToken: fakeToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
		Scope:       scope,
	}, nil
}
