# NATS Server Docker Setup

This directory contains Docker configuration for running a NATS server with JetStream enabled.

## Features

- NATS server with JetStream enabled
- HTTP monitoring on port 8222
- Persistent storage for data and configuration
- Easy deployment with Docker Compose

## Requirements

- Docker
- Docker Compose

## Usage

### Starting the NATS Server

To start the NATS server, run the following command from the `nats-docker` directory:

```bash
docker-compose up -d
```

This will start the NATS server in detached mode.

### Stopping the NATS Server

To stop the NATS server, run:

```bash
docker-compose down
```

### Viewing Logs

To view the logs of the NATS server, run:

```bash
docker-compose logs -f
```

## Connection Information

- Client connections: `localhost:4222`
- HTTP monitoring: `http://localhost:8222`
- Clustering port: `6222`

## Monitoring

You can access the HTTP monitoring interface by navigating to `http://localhost:8222` in your web browser.

## Using with Go Applications

To use this NATS server with Go applications, you can connect to it using the NATS Go client:

```go
package main

import (
    "log"
    
    "github.com/nats-io/nats.go"
)

func main() {
    // Connect to NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()
    
    // Use the connection
    // ...
}
```

Make sure to add the NATS Go client to your project:

```bash
go get github.com/nats-io/nats.go
```
