// Package models contains data structures for token requests and responses
package models

import "time"

// TokenRequest represents a request for a token
type TokenRequest struct {
	RequestID    string    `json:"request_id"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	Timestamp    time.Time `json:"timestamp"`
}

// NewTokenRequest creates a new token request
func NewTokenRequest(clientID, clientSecret string) *TokenRequest {
	return &TokenRequest{
		RequestID:    generateID(),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Timestamp:    time.Now(),
	}
}

// TokenResponse represents a response with token information
type TokenResponse struct {
	RequestID   string    `json:"request_id"`
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	Error       string    `json:"error,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewTokenResponse creates a new token response
func NewTokenResponse(requestID, accessToken, tokenType string, expiresIn int) *TokenResponse {
	return &TokenResponse{
		RequestID:   requestID,
		AccessToken: accessToken,
		TokenType:   tokenType,
		ExpiresIn:   expiresIn,
		Timestamp:   time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(requestID, errorMessage string) *TokenResponse {
	return &TokenResponse{
		RequestID: requestID,
		Error:     errorMessage,
		Timestamp: time.Now(),
	}
}
