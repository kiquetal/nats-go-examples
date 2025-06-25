// Package main provides a simple NATS subscriber example
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiquetal/nats-go-examples/internal/config"
	"github.com/kiquetal/nats-go-examples/internal/logger"
	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/kiquetal/nats-go-examples/pkg/pubsub"
	"github.com/nats-io/nats.go"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	subject := flag.String("subject", "messages", "Subject to subscribe to")
	queue := flag.String("queue", "", "Queue group name (optional)")
	flag.Parse()

	// Load configuration
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Create a logger
	log := logger.DefaultLogger("subscriber")
	log.Info("Starting NATS subscriber")

	// Create a new subscriber using the configuration
	subscriber, err := pubsub.NewSubscriber(appConfig.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS: %v", err)
	}
	defer subscriber.Close()

	log.Info("Connected to NATS at %s", appConfig.NATS.URL)
	log.Info("Subscribing to subject: %s", *subject)

	// Create message handler
	handler := func(msg *models.Message) error {
		log.Info("Received message on subject %s:", msg.Subject)
		log.Info("  ID: %s", msg.ID)
		log.Info("  Body: %s", msg.Body)
		log.Info("  Timestamp: %s", msg.Timestamp.Format(time.RFC3339))
		log.Info("  Metadata: %v", msg.Metadata)
		return nil
	}

	// Subscribe to messages
	var sub *nats.Subscription
	if *queue != "" {
		log.Info("Using queue group: %s", *queue)
		sub, err = subscriber.QueueSubscribeMessage(*subject, *queue, handler)
	} else {
		sub, err = subscriber.SubscribeMessage(*subject, handler)
	}

	if err != nil {
		log.Fatal("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Info("Subscriber started. Press Ctrl+C to exit.")

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signals
	log.Info("Received shutdown signal, exiting...")
}
