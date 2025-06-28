// Package main implements a worker that obtains tokens from an identity provider
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kiquetal/nats-go-examples/internal/config"
	"github.com/kiquetal/nats-go-examples/internal/idp"
	"github.com/kiquetal/nats-go-examples/internal/logger"
	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/nats-io/nats.go"
)

const (
	tokenSubject = "token.request"
	defaultQueue = "token-workers"
)

// createTokenRequestHandler returns a callback function for processing token requests
func createTokenRequestHandler(idpClient *idp.Client, log *logger.Logger) nats.MsgHandler {
	return func(msg *nats.Msg) {
		// Parse the token request
		var request models.TokenRequest
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			log.Error("Failed to parse token request: %v", err)
			sendErrorResponse(msg, "", "Invalid request format")
			return
		}

		log.Info("Received token request for client ID: %s (Request ID: %s)",
			request.ClientID, request.RequestID)

		// Create credentials from the request
		credentials := &idp.ClientCredentials{
			ClientID:     request.ClientID,
			ClientSecret: request.ClientSecret,
			Scope:        "openid profile",
		}

		var response *models.TokenResponse

		// Obtain token from IDP
		// For development/testing, use the simulation method
		// In production, use the real method: idpClient.GetTokenWithClientCredentials
		tokenResp, err := idpClient.GetTokenWithClientCredentials(credentials)
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
			tokenResp.Scope,
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
	}
}

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	idpURL := flag.String("idp-url", idp.DefaultBaseURL, "IDP base URL")
	idpTokenPath := flag.String("idp-token-path", idp.DefaultTokenEndpoint, "IDP token endpoint path")
	queueName := flag.String("queue", defaultQueue, "Queue group name for load balancing")
	nameSuffix := flag.String("name-suffix", "", "Suffix to append to the client name (e.g. pod name)")
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

	// Create IDP client with custom token endpoint (env vars are handled within the idp package)
	idpClient := idp.NewClient(*idpURL, idp.WithTokenEndpoint(*idpTokenPath))
	log.Info("IDP client created")

	// Create a WaitGroup to track when connection is ready
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a client name that includes the pod name if available
	clientName := "Token Worker"
	if *nameSuffix != "" {
		clientName = fmt.Sprintf("%s-%s", clientName, *nameSuffix)
	} else {
		// Try to get pod name from environment variable
		if podName := os.Getenv("POD_NAME"); podName != "" {
			clientName = fmt.Sprintf("%s-%s", clientName, podName)
		} else if hostname, err := os.Hostname(); err == nil {
			// Fall back to hostname if pod name is not available
			clientName = fmt.Sprintf("%s-%s", clientName, hostname)
		}
	}

	// Configure connection options
	opts := []nats.Option{
		nats.Name(clientName),               // Set client name with unique identifier
		nats.ReconnectWait(5 * time.Second), // Wait 5 seconds between reconnect attempts
		nats.MaxReconnects(10),              // Try to reconnect up to 10 times
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warn("Disconnected from NATS: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("Reconnected to NATS server at %s", nc.ConnectedUrl())
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Error("NATS error: %v", err)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Warn("NATS connection closed")
		}),
		// The most important handler - signals when the connection is established
		nats.ConnectHandler(func(nc *nats.Conn) {
			log.Info("Connected to NATS at %s", nc.ConnectedUrl())
			// Signal that we're connected
			wg.Done()
		}),
	}

	// Connect to NATS with options
	log.Info("Connecting to NATS at %s...", appConfig.NATS.URL)
	natsConn, err := nats.Connect(appConfig.NATS.URL, opts...)
	if err != nil {
		log.Fatal("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	// Wait for the connection to be established
	wg.Wait()
	log.Info("NATS connection established successfully")

	log.Info("Subscribing to token requests on %s with queue group %s", tokenSubject, *queueName)

	// Create the token request handler and subscribe to the token subject with queue group
	handler := createTokenRequestHandler(idpClient, log)
	_, err = natsConn.QueueSubscribe(tokenSubject, *queueName, handler)
	if err != nil {
		log.Fatal("Failed to subscribe to token requests: %v", err)
	}

	log.Info("Token worker is running in queue group %s. Press Ctrl+C to exit.", *queueName)

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
