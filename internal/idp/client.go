// Package idp provides functionality for interacting with identity providers
package idp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	baseURL    string
	httpClient *http.Client
}

// ClientCredentials holds the credentials for a client
type ClientCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// NewClient creates a new IDP client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetTokenWithClientCredentials obtains a token using client credentials
func (c *Client) GetTokenWithClientCredentials(credentials *ClientCredentials) (*TokenResponse, error) {
	// Create form data
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", credentials.ClientID)
	formData.Set("client_secret", credentials.ClientSecret)

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", c.baseURL),
		strings.NewReader(formData.Encode()))
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

	return &TokenResponse{
		AccessToken: fakeToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
	}, nil
}
