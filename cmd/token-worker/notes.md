# Kubernetes Deployment Guide for Token Worker

This guide explains how to deploy the token worker service on Kubernetes, with special attention to handling instance naming for multiple replicas.

## Naming Concerns with Multiple Replicas

The token worker uses NATS client names and queue groups that must be properly configured for Kubernetes deployments:

1. **Client Names**: Each NATS connection has a name set via `nats.Name("Token Worker")` 
2. **Queue Groups**: Workers use a queue group (default: "token-workers") for load balancing

When running multiple replicas in Kubernetes, having unique client names improves:
- Monitoring and observability
- Troubleshooting connection issues
- Proper client tracking in NATS server

## Running Locally

### Using Command Line Flags

```bash
# Run with default settings
go run cmd/token-worker/main.go

# Run with custom configuration file
go run cmd/token-worker/main.go -config configs/custom.json

# Run with specified queue group and name suffix
go run cmd/token-worker/main.go -queue custom-workers -name-suffix worker1

# Run with IDP URL specified
go run cmd/token-worker/main.go -idp-url https://my-idp.example.com
```

### Using Environment Variables

```bash
# Run with environment variables
NATS_URL=nats://localhost:4222 QUEUE_GROUP=token-workers WORKER_NAME_SUFFIX=worker1 go run cmd/token-worker/main.go

# Run multiple workers with different identifiers (in separate terminals)
NATS_URL=nats://localhost:4222 WORKER_NAME_SUFFIX=worker1 LOG_LEVEL=debug IDP_URL=https://idp.example.com IDP_TOKEN_PATH="/realms/phoenix/protocol/openid-connect/token" go run cmd/token-worker/main.go
NATS_URL=nats://localhost:4222 WORKER_NAME_SUFFIX=worker2 LOG_LEVEL=debug IDP_URL=https://idp.example.com IDP_TOKEN_PATH="/realms/phoenix/protocol/openid-connect/token" go run cmd/token-worker/main.go

# Run with multiple environment variables
NATS_URL=nats://localhost:4222 \
QUEUE_GROUP=token-workers \
WORKER_NAME_SUFFIX=worker1 \
LOG_LEVEL=debug \
TOKEN_SUBJECT=token.request \
IDP_URL=https://idp.example.com \
IDP_TOKEN_PATH="/realms/phoenix/protocol/openid-connect/token" \
go run cmd/token-worker/main.go
```

Available environment variables:
- `NATS_URL`: NATS server URL (default: nats://localhost:4222)
- `QUEUE_GROUP`: NATS queue group name (default: token-workers)
- `WORKER_NAME_SUFFIX`: Suffix to append to the worker name for uniqueness
- `LOG_LEVEL`: Logging level (debug, info, warn, error) (default: info)
- `TOKEN_SUBJECT`: NATS subject for token requests (default: token.request)
- `MAX_RECONNECT`: Maximum reconnect attempts (default: 10)
- `RECONNECT_WAIT`: Wait time between reconnects in seconds (default: 5)
- `IDP_URL`: URL of the Identity Provider service (default: https://idp.example.com)
- `IDP_TOKEN_PATH`: Path to token endpoint (default: /realms/phoenix/protocol/openid-connect/token)
- `POD_NAME`: Pod name for Kubernetes deployments (used for unique client naming)

## Deployment Steps

### 1. Create a ConfigMap for Application Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: token-worker-config
data:
  app.json: |
    {
      "environment": "production",
      "logLevel": "info",
      "nats": {
        "url": "nats://nats-cluster.nats.svc.cluster.local:4222",
        "allowReconnect": true,
        "maxReconnect": 10,
        "reconnectWait": 5
      },
      "idp": {
        "url": "https://idp.example.com",
        "timeout": 30
      }
    }
```

### 2. Create a Deployment for Token Worker

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: token-worker
  labels:
    app: token-worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: token-worker
  template:
    metadata:
      labels:
        app: token-worker
    spec:
      containers:
      - name: token-worker
        image: token-worker:latest
        imagePullPolicy: Always
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: IDP_URL
          value: "https://idp.example.com"
        command:
        - "/app/token-worker"
        args:
        - "-config"
        - "/app/configs/app.json"
        - "-queue"
        - "token-workers"
        - "-name-suffix"
        - "$(POD_NAME)"
        - "-idp-url"
        - "$(IDP_URL)"
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
      volumes:
      - name: config-volume
        configMap:
          name: token-worker-config
```

### 3. Modify the Token Worker Code

For the token worker to use the pod name for uniqueness, we need to add a new flag that incorporates the pod name into the NATS client name:

```go
// Add a new flag
nameSuffix := flag.String("name-suffix", "", "Suffix to append to the client name")

// Modify the nats.Name option
clientName := "Token Worker"
if *nameSuffix != "" {
    clientName = fmt.Sprintf("%s-%s", clientName, *nameSuffix)
}

// Use in options
opts := []nats.Option{
    nats.Name(clientName),
    // ...other options...
}
```

## Implementation Notes

1. **Pod Name as Unique Identifier**:
   - The deployment injects the Kubernetes pod name into each container via the downward API
   - This name is used as a suffix to create unique NATS client names

2. **Queue Group Consistency**:
   - All workers should use the same queue group name for proper load balancing
   - The queue name is explicitly set via command-line arguments

3. **Client Name Format**:
   - Final NATS client names will appear as: "Token Worker-token-worker-5d4f9b8c7d"
   - This format makes it easy to identify which pod corresponds to which NATS client

4. **IDP Configuration**:
   - The Identity Provider URL can be configured via env var or command-line flag
   - In production, consider using secure methods to provide the IDP URL

## Monitoring

To verify each worker has a unique client name:

```bash
# Get the list of connected clients from NATS server
kubectl exec -it nats-cluster-0 -- nats-top --server http://localhost:8222
```

## Scaling

When scaling the deployment, each new pod will automatically get a unique name:

```bash
# Scale to more replicas
kubectl scale deployment token-worker --replicas=5
```

## Best Practices

1. Always include the pod name or another unique identifier in the NATS client name
2. Use consistent queue group names across all replicas
3. Set reasonable reconnect settings for Kubernetes environment
4. Consider using Pod Disruption Budgets (PDB) to maintain service during cluster operations
5. Properly configure IDP timeout settings based on expected response times
