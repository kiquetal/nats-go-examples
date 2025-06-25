# NATS with Go Examples

This repository contains examples of using NATS messaging system with Go programming language.

## Overview

[NATS](https://nats.io/) is a simple, secure and high-performance open source messaging system designed for distributed systems. This project demonstrates how to use NATS with Go for various messaging patterns.

## Project Structure

```
├── go.mod                 # Go module definition
├── README.md              # Project documentation
├── api/                   # API definitions
├── cmd/                   # Application entry points
│   ├── publisher/         # Publisher executable
│   └── subscriber/        # Subscriber executable
├── configs/               # Configuration files
│   └── app.json           # Example application config
├── docs/                  # Documentation files
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   └── logger/            # Logging functionality
├── nats-docker/           # Docker setup for NATS server
│   ├── docker-compose.yml # Docker Compose configuration
│   └── Dockerfile         # NATS server Dockerfile
├── pkg/                   # Public library code
│   ├── models/            # Shared data models
│   └── pubsub/            # NATS pub/sub functionality
└── scripts/               # Utility scripts
```

## Prerequisites

- Go 1.20+
- Docker and Docker Compose (for running NATS server locally)

## Getting Started

### 1. Start the NATS Server

```bash
cd nats-docker
docker-compose up -d
```

This will start a NATS server accessible at `localhost:4222` with monitoring at `localhost:8222`.

### 2. Run the Subscriber

In one terminal, start the subscriber to receive messages:

```bash
# Using default configuration
go run cmd/subscriber/main.go

# Using custom configuration file
go run cmd/subscriber/main.go -config configs/app.json

# Specifying subject and queue group
go run cmd/subscriber/main.go -subject orders.new -queue order-processors
```

### 3. Run the Publisher

In another terminal, run the publisher to send messages:

```bash
# Using default configuration
go run cmd/publisher/main.go

# Using custom configuration file
go run cmd/publisher/main.go -config configs/app.json

# Customizing subject and publish interval (milliseconds)
go run cmd/publisher/main.go -subject orders.new -interval 2000
```

### 4. Configuration Options

You can configure the applications using:

1. **Config files**: JSON files in the `configs/` directory
2. **Command-line flags**:
   - `-config`: Path to config file
   - `-subject`: Subject to publish/subscribe to
   - `-interval`: Publishing interval (publisher only)
   - `-queue`: Queue group name (subscriber only)
3. **Environment variables**:
   - `NATS_URL`: NATS server URL
   - `APP_ENV`: Application environment (dev, test, prod)
   - `APP_LOG_LEVEL`: Log level (debug, info, warn, error)

## Code Examples

### Publisher Example

```go
// Create a publisher
publisher, err := pubsub.NewPublisher("nats://localhost:4222")
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer publisher.Close()

// Create and publish a message
msg := models.NewMessage("greetings", "Hello, NATS!")
msg.AddMetadata("sender", "example")
if err := publisher.PublishMessage(msg); err != nil {
    log.Fatalf("Failed to publish: %v", err)
}
```

### Subscriber Example

```go
// Create a subscriber
subscriber, err := pubsub.NewSubscriber("nats://localhost:4222")
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer subscriber.Close()

// Define a message handler
handler := func(msg *models.Message) error {
    fmt.Printf("Received: %s\n", msg.Body)
    return nil
}

// Subscribe to messages
sub, err := subscriber.SubscribeMessage("greetings", handler)
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}
defer sub.Unsubscribe()

// Wait for messages
select {}
```

## Key Concepts Demonstrated

- Simple Publish/Subscribe
- Queue Groups for load balancing
- Configuration management
- Structured logging
- Graceful shutdown handling

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
