# NATS with Go Examples

This repository contains examples of using NATS messaging system with Go programming language.

## Overview

[NATS](https://nats.io/) is a simple, secure and high-performance open source messaging system designed for distributed systems. This project demonstrates how to use NATS with Go for various messaging patterns.

## Project Structure

```
├── go.mod
├── nats-docker/           # Docker setup for NATS server
├── publisher/             # Example publisher implementations
├── subscriber/            # Example subscriber implementations
├── models/                # Data models shared across packages
└── examples/              # Complete example scenarios
```

## Prerequisites

- Go 1.24+
- Docker and Docker Compose (for running NATS server locally)

## Getting Started

### 1. Start the NATS Server

```bash
cd nats-docker
docker-compose up -d
```

This will start a NATS server accessible at `localhost:4222` with monitoring at `localhost:8222`.

### 2. Run Examples

See individual examples in the `examples` directory. Each example has its own README with instructions.

Basic usage pattern:

```go
// Example of a simple NATS publisher
package main

import (
    "github.com/kiquetal/nats-go-example/publisher"
    "log"
)

func main() {
    p, err := publisher.NewSimplePublisher("nats://localhost:4222")
    if err != nil {
        log.Fatalf("Failed to create publisher: %v", err)
    }
    defer p.Close()
    
    if err := p.Publish("greetings", []byte("Hello, NATS!")); err != nil {
        log.Fatalf("Failed to publish message: %v", err)
    }
}
```

## Key Concepts Demonstrated

- Simple Publish/Subscribe
- Request/Reply pattern
- Queue Groups
- JetStream persistence
- Subject-based routing
- Wildcards in subjects

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
