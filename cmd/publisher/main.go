// Package main provides a simple NATS publisher example
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiquetal/nats-go-examples/internal/config"
	"github.com/kiquetal/nats-go-examples/internal/logger"
	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/kiquetal/nats-go-examples/pkg/pubsub"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	subject := flag.String("subject", "messages", "Subject to publish to")
	interval := flag.Int("interval", 1000, "Publish interval in milliseconds")
	flag.Parse()

	// Load configuration
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a logger
	log := logger.DefaultLogger("publisher")
	log.Info("Starting NATS publisher")

	// Create a new publisher using the configuration
	publisher, err := pubsub.NewPublisher(appConfig.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS: %v", err)
	}
	defer publisher.Close()

	log.Info("Connected to NATS at %s", appConfig.NATS.URL)
	log.Info("Publishing to subject: %s", *subject)
	log.Info("Publishing interval: %d ms", *interval)

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for regular publishing
	ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
	defer ticker.Stop()

	count := 0
	running := true

	for running {
		select {
		case <-ticker.C:
			count++
			// Create a message
			msg := models.NewMessage(*subject, fmt.Sprintf("Message #%d", count))
			msg.AddMetadata("publisher", "example")
			msg.AddMetadata("timestamp", time.Now().Format(time.RFC3339))
			msg.AddMetadata("environment", appConfig.Environment)

			// Publish the message
			if err := publisher.PublishMessage(msg); err != nil {
				log.Error("Error publishing message: %v", err)
				continue
			}

			log.Info("Published message #%d to %s", count, *subject)

		case <-signals:
			log.Info("Received shutdown signal, exiting...")
			running = false
		}
	}

	log.Info("Publisher shutdown complete")
}
