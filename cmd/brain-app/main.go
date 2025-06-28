// Package main implements the brain-app HTTP server
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiquetal/nats-go-examples/internal/cache"
	"github.com/kiquetal/nats-go-examples/internal/config"
	"github.com/kiquetal/nats-go-examples/internal/logger"
	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/nats-io/nats.go"
)

const (
	tokenSubject    = "token.request"
	defaultTokenTTL = 55 * time.Minute // Cache tokens for slightly less than their typical 1-hour expiry
)

// TokenServer handles token requests via HTTP and NATS
type TokenServer struct {
	natsConn       *nats.Conn
	tokenCache     *cache.TokenCache
	log            *logger.Logger
	requestTimeout time.Duration
}

// ClientCredentialsRequest represents a request for client credentials
type ClientCredentialsRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	port := flag.Int("port", 8080, "HTTP server port")
	requestTimeout := flag.Int("request-timeout", 5, "Timeout for NATS requests in seconds")
	flag.Parse()

	// Load configuration
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	log := logger.DefaultLogger("brain-app")
	log.Info("Starting brain-app server")

	// Create token cache
	tokenCache := cache.NewTokenCache()
	log.Info("Token cache initialized")

	// Connect to NATS
	natsConn, err := nats.Connect(appConfig.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()
	log.Info("Connected to NATS at %s", appConfig.NATS.URL)

	// Create token server
	server := &TokenServer{
		natsConn:       natsConn,
		tokenCache:     tokenCache,
		log:            log,
		requestTimeout: time.Duration(*requestTimeout) * time.Second,
	}

	// Set up HTTP routes
	http.HandleFunc("/token", server.handleTokenRequest)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server in a goroutine
	go func() {
		serverAddr := fmt.Sprintf(":%d", *port)
		log.Info("Starting HTTP server on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatal("HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	log.Info("Received shutdown signal, exiting...")
}

// handleTokenRequest processes HTTP requests for tokens
func (s *TokenServer) handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for query param to skip cache
	skipCache := false
	if v := r.URL.Query().Get("skip_cache"); v == "1" || v == "true" {
		skipCache = true
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		s.log.Error("Failed to read request body: %v", err)
		return
	}
	defer r.Body.Close()

	// Parse client credentials
	var creds ClientCredentialsRequest
	if err := json.Unmarshal(body, &creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		s.log.Error("Failed to parse request: %v", err)
		return
	}

	// Validate client credentials
	if creds.ClientID == "" || creds.ClientSecret == "" {
		http.Error(w, "Client ID and Client Secret are required", http.StatusBadRequest)
		return
	}

	// Check cache first, unless skipCache is set
	if !skipCache {
		if token, found := s.tokenCache.Get(creds.ClientID); found {
			s.log.Info("Serving cached token for client ID: %s", creds.ClientID)

			// Return cached token
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"access_token": token,
				"token_type":   "Bearer",
				"source":       "cache",
			})
			return
		}
	}

	// Create token request
	tokenReq := models.NewTokenRequest(creds.ClientID, creds.ClientSecret)

	// Convert request to JSON
	reqData, err := json.Marshal(tokenReq)
	if err != nil {
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		s.log.Error("Failed to marshal token request: %v", err)
		return
	}

	// Send request to NATS and wait for response with timeout
	s.log.Info("Sending token request for client ID: %s (Request ID: %s)",
		creds.ClientID, tokenReq.RequestID)

	msg, err := s.natsConn.Request(tokenSubject, reqData, s.requestTimeout)
	if err != nil {
		if err == nats.ErrTimeout {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			s.log.Error("Token request timed out for request ID: %s", tokenReq.RequestID)
		} else {
			http.Error(w, "Failed to process request", http.StatusInternalServerError)
			s.log.Error("Failed to send token request: %v", err)
		}
		return
	}

	// Parse the response
	var response models.TokenResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
		s.log.Error("Failed to parse token response: %v", err)
		return
	}

	// Check for error in response
	if response.Error != "" {
		http.Error(w, response.Error, http.StatusBadRequest)
		s.log.Error("Token request failed: %s", response.Error)
		return
	}

	// Cache the token for future use, unless skipCache is set
	if !skipCache {
		s.tokenCache.Set(creds.ClientID, response.AccessToken, defaultTokenTTL)
		s.log.Info("Token cached for client ID: %s", creds.ClientID)
	}

	// Return token to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": response.AccessToken,
		"token_type":   response.TokenType,
		"source":       "idp",
	})
}
