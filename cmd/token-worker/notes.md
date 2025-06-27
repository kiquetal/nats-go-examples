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
        command:
        - "/app/token-worker"
        args:
        - "-config"
        - "/app/configs/app.json"
        - "-queue"
        - "token-workers"
        - "-name-suffix"
        - "$(POD_NAME)"
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
