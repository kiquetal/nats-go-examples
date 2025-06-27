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
│   ├── subscriber/        # Subscriber executable
│   └── brain-app/         # Token management service
├── configs/               # Configuration files
│   └── app.json           # Example application config
├── docs/                  # Documentation files
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   ├── logger/            # Logging functionality
│   └── cache/             # Token caching
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

# Using environment variables
NATS_URL=nats://localhost:4222 APP_ENV=dev APP_LOG_LEVEL=debug go run cmd/subscriber/main.go
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

# Using environment variables
NATS_URL=nats://localhost:4222 APP_ENV=dev APP_LOG_LEVEL=debug go run cmd/publisher/main.go
```

### 4. Run the Brain App

The brain app serves as a token management service that uses NATS to communicate with token workers:

```bash
# Using default settings
go run cmd/brain-app/main.go

# Specifying port and request timeout
go run cmd/brain-app/main.go -port 8080 -request-timeout 5

# Using environment variables
NATS_URL=nats://localhost:4222 PORT=8080 REQUEST_TIMEOUT=5 go run cmd/brain-app/main.go

# Using a custom configuration file
go run cmd/brain-app/main.go -config configs/brain-app.json
```

### 5. Configuration Options

You can configure the applications using:

1. **Config files**: JSON files in the `configs/` directory
2. **Command-line flags**:
   - `-config`: Path to config file
   - `-subject`: Subject to publish/subscribe to
   - `-interval`: Publishing interval (publisher only)
   - `-queue`: Queue group name (subscriber only)
   - `-port`: HTTP port (brain-app only)
   - `-request-timeout`: NATS request timeout in seconds (brain-app only)
3. **Environment variables**:
   - `NATS_URL`: NATS server URL
   - `APP_ENV`: Application environment (dev, test, prod)
   - `APP_LOG_LEVEL`: Log level (debug, info, warn, error)
   - `PORT`: HTTP server port (brain-app only)
   - `REQUEST_TIMEOUT`: NATS request timeout in seconds (brain-app only)

## Running with Docker

### 1. Building Docker Images

From the root directory of the project, build the Docker images for the publisher and subscriber:

```bash
# Build publisher image
docker build -t nats-publisher -f cmd/publisher/Dockerfile .

# Build subscriber image
docker build -t nats-subscriber -f cmd/subscriber/Dockerfile .

# Build brain-app image
docker build -t brain-app -f cmd/brain-app/Dockerfile .
```

### 2. Running the Containers

Make sure your NATS server is running first, then start the containers:

```bash
# Run subscriber container
docker run --network host -e NATS_URL=nats://localhost:4222 nats-subscriber

# Run publisher container
docker run --network host -e NATS_URL=nats://localhost:4222 nats-publisher

# Run brain-app container
docker run --network host -e NATS_URL=nats://localhost:4222 -e PORT=8080 brain-app
```

### 3. Docker Compose (Optional)

You can also use Docker Compose to run the entire system. Create a docker-compose.yml file at the root of the project:

```yaml
version: '3'
services:
  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "8222:8222"
    
  publisher:
    build:
      context: .
      dockerfile: cmd/publisher/Dockerfile
    environment:
      - NATS_URL=nats://nats:4222
    depends_on:
      - nats
    
  subscriber:
    build:
      context: .
      dockerfile: cmd/subscriber/Dockerfile
    environment:
      - NATS_URL=nats://nats:4222
    depends_on:
      - nats
      
  brain-app:
    build:
      context: .
      dockerfile: cmd/brain-app/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - NATS_URL=nats://nats:4222
      - PORT=8080
      - REQUEST_TIMEOUT=5
    depends_on:
      - nats
```

Then run:

```bash
docker-compose up -d
```

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

### Brain App Token Request Example

```bash
# Request a token using curl
curl -X POST http://localhost:8080/token \
  -H 'Content-Type: application/json' \
  -d '{
    "client_id": "example-client",
    "client_secret": "example-secret"
  }'
```

## Key Concepts Demonstrated

- Simple Publish/Subscribe
- Queue Groups for load balancing
- Request-Reply pattern for token service
- Token caching to reduce NATS traffic
- Configuration management
- Structured logging
- Graceful shutdown handling

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
