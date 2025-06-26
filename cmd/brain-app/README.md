# Brain App

## Overview

The Brain App is a service that works as an intelligent proxy for obtaining tokens from an identity provider (IDP). It implements caching to minimize external calls and uses NATS for communication with worker services.

## Architecture

```
                                      ┌─────────────────┐
                                      │                 │
                                      │  Identity       │
                                      │  Provider (IDP) │
                                      │                 │
                                      └────────┬────────┘
                                               │
                                               │ HTTP
                                               │
┌─────────────┐   HTTP    ┌─────────────┐     │     ┌─────────────┐
│             │           │             │     │     │             │
│   Client    ├──────────►│  Brain App  │     │     │ Token       │
│             │           │             │◄────┴────►│ Worker #1   │
└─────────────┘           │  (Caching)  │           │             │
                          │             │           └─────────────┘
                          └──────┬──────┘
                                 │                  ┌─────────────┐
                                 │                  │             │
                                 │                  │ Token       │
                                 │  NATS           │ Worker #2   │
                                 └─────────────────►│             │
                                    Request/Reply   └─────────────┘
                                    (Load Balanced)
                                                   ┌─────────────┐
                                                   │             │
                                                   │ Token       │
                                                   │ Worker #3   │
                                                   │             │
                                                   └─────────────┘
```

## How It Works

1. **Client Request**: A client sends an HTTP request to the Brain App with client credentials (client_id and client_secret)

2. **Cache Check**: The Brain App checks its internal cache for an existing valid token
   - If a valid token exists, it is returned immediately to the client

3. **NATS Request**: If no cached token is found, the Brain App sends a request to NATS
   - Using the request-reply pattern for synchronous communication
   - The request includes the client credentials

4. **Worker Processing**: One of the token workers (selected via queue group) processes the request
   - The worker calls the Identity Provider to obtain a token
   - The worker sends the token back to the Brain App via NATS

5. **Response Handling**: The Brain App receives the token response
   - The token is cached for future requests (configurable TTL, default 55 minutes)
   - The token is returned to the client

6. **Error Handling**: If any errors occur (timeout, IDP error), an appropriate error response is sent to the client

## Key Features

- **Token Caching**: Minimizes external calls to the Identity Provider
- **High Availability**: Multiple token workers can be deployed in a queue group
- **Load Balancing**: NATS distributes requests across available workers
- **Synchronous Client Experience**: Returns a synchronous response despite asynchronous backend
- **Configurable**: Timeout, cache TTL, and other parameters can be configured

## Configuration Options

- `-config`: Path to configuration file
- `-port`: HTTP server port (default: 8080)
- `-request-timeout`: Timeout for NATS requests in seconds (default: 5)

## Example Usage

```bash
# Start the Brain App with default settings
go run cmd/brain-app/main.go

# Start with custom configuration
go run cmd/brain-app/main.go -config configs/custom.json -port 9090 -request-timeout 10
```

## API Endpoints

### GET /health

Health check endpoint that returns HTTP 200 if the service is running.

### POST /token

Endpoint for requesting tokens.

**Request Body**:
```json
{
  "client_id": "my-client",
  "client_secret": "my-secret"
}
```

**Success Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "source": "cache" // or "idp" if freshly obtained
}
```

**Error Response** (various HTTP error codes):
```json
{
  "error": "Error message"
}
```

## Docker Deployment

```bash
# Build the Docker image
docker build -t brain-app:latest -f cmd/brain-app/Dockerfile .

# Run the container
docker run -p 8080:8080 -e NATS_URL=nats://nats-server:4222 brain-app:latest
```

## Production Considerations

- Configure appropriate cache TTL based on token expiration times
- Monitor cache hit rate to optimize performance
- Scale token workers horizontally for high availability
- Configure appropriate timeouts for both HTTP and NATS requests
