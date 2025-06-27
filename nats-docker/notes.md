# NATS Server on Kubernetes - Best Practices

This document outlines the best practices for running the latest version of NATS (2.10.x as of June 2025) on Kubernetes.

## NATS Version Recommendations

- Use NATS server version 2.10.0 or later for the latest features, performance improvements, and security updates
- Use the official `nats:2.10-alpine` image for minimal container size
- Consider `-scratch` variant images for even smaller footprint in production environments
- Always specify a specific version tag rather than `:latest` for production deployments

## Kubernetes Deployment Strategies

### Basic Deployment

For simple use cases without persistence or clustering:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nats-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nats-server
  template:
    metadata:
      labels:
        app: nats-server
    spec:
      containers:
      - name: nats
        image: nats:2.10-alpine
        ports:
        - containerPort: 4222  # Client connections
        - containerPort: 8222  # HTTP monitoring
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
```

### StatefulSet for Clustering

For production-grade NATS clusters:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nats-cluster
spec:
  serviceName: "nats"
  replicas: 3
  selector:
    matchLabels:
      app: nats
  template:
    metadata:
      labels:
        app: nats
    spec:
      containers:
      - name: nats
        image: nats:2.10-alpine
        command:
        - "nats-server"
        - "--name=nats-$(POD_NAME)"
        - "--cluster=nats://0.0.0.0:6222"
        - "--cluster_routes=nats://nats-0.nats:6222,nats://nats-1.nats:6222,nats://nats-2.nats:6222"
        - "--http_port=8222"
        - "--jetstream"
        - "--store_dir=/data"
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        volumeMounts:
        - name: data
          mountPath: /data
        ports:
        - containerPort: 4222  # Client
        - containerPort: 6222  # Clustering
        - containerPort: 8222  # HTTP
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 1
            memory: 1Gi
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 5Gi
```

## Operator-Based Installation (Recommended)

The NATS Operator simplifies deploying and managing NATS clusters:

```bash
# Install the NATS Operator
kubectl apply -f https://github.com/nats-io/k8s/releases/download/v0.19.0/nats-operator.yaml

# Create a NATS cluster with JetStream
kubectl apply -f - << EOF
apiVersion: nats.io/v1alpha2
kind: NatsCluster
metadata:
  name: nats-cluster
spec:
  size: 3
  version: "2.10.0"
  jetstream:
    enabled: true
    memStorage:
      enabled: false
    fileStorage:
      enabled: true
      storageDirectory: /data/jetstream
      storageClassname: standard
      size: 10Gi
EOF
```

## Helm Chart Installation (Recommended)

The NATS Helm charts provide the most flexible and maintainable way to deploy NATS on Kubernetes.

### Adding the NATS Helm Repository

```bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
```

### Basic NATS Server Installation

```bash
helm install my-nats nats/nats --version 0.22.0 \
  --set jetstream.enabled=true \
  --set jetstream.fileStorage.enabled=true \
  --set jetstream.fileStorage.size=10Gi
```

### Production-Ready Installation

For production environments with clustering, monitoring and security:

```bash
helm install nats nats/nats --version 0.22.0 -f - <<EOF
global:
  image:
    registry: docker.io
    repository: nats
    tag: "2.10.0-alpine"
    
cluster:
  enabled: true
  replicas: 3
  
jetstream:
  enabled: true
  memStorage:
    enabled: true
    size: 2Gi
  fileStorage:
    enabled: true
    size: 10Gi
    storageClassName: ssd

auth:
  enabled: true
  basic:
    users:
      - user: admin
        password: admin_password
        permissions:
          publish: ">"
          subscribe: ">"
      - user: token_service
        password: service_password
        permissions:
          publish: "token.*"
          subscribe: "token.*"
          
tls:
  enabled: true
  secretName: nats-server-tls
  
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1
    memory: 1Gi
    
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    
prometheus:
  enabled: true
  
grafana:
  enabled: true
  dashboards:
    enabled: true
    
podDisruptionBudget:
  enabled: true
  minAvailable: 2
EOF
```

### Customizing Storage for JetStream

For optimized JetStream performance with specific storage class:

```bash
helm install nats nats/nats -f - <<EOF
nats:
  jetstream:
    enabled: true
    fileStorage:
      enabled: true
      size: 50Gi
      storageClassName: fast-ssd
      storageDirectory: /data/jetstream
EOF
```

### Providing TLS Certificates

```bash
# Create TLS secrets first
kubectl create secret tls nats-server-tls \
  --cert=/path/to/tls.crt \
  --key=/path/to/tls.key

# Then reference in Helm chart
helm install nats nats/nats --set nats.tls.enabled=true --set nats.tls.secretName=nats-server-tls
```

### Upgrading NATS Helm Deployment

```bash
helm repo update
helm upgrade nats nats/nats --version 0.22.0 -f values.yaml
```

### Installing with Custom Values File

Create a `values.yaml` file with your configuration:

```yaml
nats:
  image: nats:2.10.0-alpine
  jetstream:
    enabled: true
    fileStorage:
      enabled: true
      size: 10Gi
  auth:
    enabled: true
    basic:
      users:
        - user: app
          password: password
```

Then install using:

```bash
helm install nats nats/nats -f values.yaml
```

## Security Best Practices

1. **Enable TLS for Client and Cluster connections**:
```yaml
spec:
  tls:
    serverSecret: "nats-server-tls"
    routesSecret: "nats-routes-tls"
```

2. **Use Authentication**:
```yaml
spec:
  auth:
    enableServiceAccounts: true
```

3. **Configure Account Resolver**:
```yaml
spec:
  accountResolver:
    type: URL
    store:
      dir: /accounts/jwt
    resolverPreload:
      ACXZC3BNDOABESL8RGOM3JKPR2C6OEIMBSFXJIBTN6IMXL2J2ZPQKJBJ: eyJ0e...
```

## Resource Management

- Start with resource requests and limits appropriate for your workload
- For production clusters, allocate at least 1Gi of memory per node
- Use horizontal pod autoscaler for NATS clients but not for the core NATS server cluster
- For JetStream, allocate sufficient storage based on message retention needs

## Monitoring

1. **Enable Prometheus Metrics**:
```yaml
spec:
  pod:
    enableMetrics: true
    metricsPort: 7777
```

2. **Install Grafana Dashboards**:
```bash
kubectl apply -f https://raw.githubusercontent.com/nats-io/k8s/main/nats-server-grafana-dashboard.json
```

## High Availability Configuration

- Use Pod Anti-affinity to distribute NATS servers across nodes:
```yaml
spec:
  template:
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - topologyKey: kubernetes.io/hostname
            labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nats
```

- Use Pod Disruption Budgets to maintain quorum during upgrades:
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: nats-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: nats
```

## JetStream Best Practices

1. **Storage Configuration**:
   - Use SSDs for JetStream storage
   - Configure appropriate storage class for your cluster
   - Use separate storage for JetStream data

2. **Limits and Quotas**:
   - Configure memory, storage, and streams/consumer limits
   - Set appropriate retention policies

```yaml
jetStream:
  memoryStorage:
    enabled: true
    size: 2Gi
  fileStorage:
    enabled: true
    storageClassName: ssd
    size: 10Gi
  maxMemory: 2Gi
  maxStorage: 10Gi
```

3. **Persistent Storage Backup**:
   - Implement a regular backup strategy for JetStream data
   - Use volume snapshots for backup

## Upgrade Strategy

- Always check the release notes for breaking changes
- Use a rolling update strategy with maxUnavailable set to 1
- Upgrade one node at a time to maintain cluster availability
- Test upgrades in a staging environment first

## Additional Resources

- [Official NATS Kubernetes Documentation](https://docs.nats.io/running-a-nats-service/nats-on-kubernetes)
- [NATS Kubernetes GitHub Repository](https://github.com/nats-io/k8s)
- [NATS Operator Documentation](https://docs.nats.io/running-a-nats-service/nats-kubernetes/nats-kubernetes-operator)
