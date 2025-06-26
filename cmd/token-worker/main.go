// Package main implements a worker that obtains tokens from an identity provider
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kiquetal/nats-go-examples/internal/config"
	"github.com/kiquetal/nats-go-examples/internal/idp"
	"github.com/kiquetal/nats-go-examples/internal/logger"
	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/nats-io/nats.go"
)

const (
	tokenSubject = "token.request"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	idpURL := flag.String("idp-url", "https://idp.example.com", "IDP base URL")
	flag.Parse()

	// Load configuration
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	log := logger.DefaultLogger("token-worker")
	log.Info("Starting token worker")

	// Create IDP client
	idpClient := idp.NewClient(*idpURL)
	log.Info("IDP client created for %s", *idpURL)

	// Connect to NATS
	natsConn, err := nats.Connect(appConfig.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	log.Info("Connected to NATS at %s", appConfig.NATS.URL)
	log.Info("Subscribing to token requests on %s", tokenSubject)

	// Subscribe to token requests
	_, err = natsConn.Subscribe(tokenSubject, func(msg *nats.Msg) {
		// Parse the token request
		var request models.TokenRequest
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			log.Error("Failed to parse token request: %v", err)
			sendErrorResponse(msg, request.RequestID, "Invalid request format")
			return
		}

		log.Info("Received token request for client ID: %s (Request ID: %s)",
			request.ClientID, request.RequestID)

		// Create credentials from the request
		credentials := &idp.ClientCredentials{
			ClientID:     request.ClientID,
			ClientSecret: request.ClientSecret,
		}

		var response *models.TokenResponse

		// Obtain token from IDP
		// For development/testing, use the simulation method
		// In production, use the real method: idpClient.GetTokenWithClientCredentials
		tokenResp, err := idpClient.SimulateTokenRetrieval(credentials)
		if err != nil {
			log.Error("Failed to obtain token: %v", err)
			sendErrorResponse(msg, request.RequestID, err.Error())
			return
		}

		log.Info("Token obtained for client ID: %s", request.ClientID)
		response = models.NewTokenResponse(
			request.RequestID,
			tokenResp.AccessToken,
			tokenResp.TokenType,
			tokenResp.ExpiresIn,
		)

		// Marshal the response
		respData, err := json.Marshal(response)
		if err != nil {
			log.Error("Failed to marshal token response: %v", err)
			sendErrorResponse(msg, request.RequestID, "Internal server error")
			return
		}

		// Reply to the request
		if err := msg.Respond(respData); err != nil {
			log.Error("Failed to send response: %v", err)
			return
		}

		log.Info("Sent token response for request ID: %s", request.RequestID)
	})

	if err != nil {
		log.Fatal("Failed to subscribe to token requests: %v", err)
	}

	log.Info("Token worker is running. Press Ctrl+C to exit.")

	// Wait for termination signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	log.Info("Received shutdown signal, exiting...")
}

// sendErrorResponse sends an error response back to the requester
func sendErrorResponse(msg *nats.Msg, requestID, errorMessage string) {
	response := models.NewErrorResponse(requestID, errorMessage)
	respData, err := json.Marshal(response)
	if err != nil {
		// Just log, can't do much else here
		return
	}
	msg.Respond(respData)
}
