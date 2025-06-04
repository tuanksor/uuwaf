# UUWAF Operator

A Kubernetes operator that monitors the UUWAF audit database and reloads configurations when changes are detected.

## Features

- Real-time monitoring of UUWAF audit database
- Automatic reload of configurations on Kubernetes pods
- Structured logging with zap
- Prometheus metrics for monitoring
- Graceful shutdown handling
- Connection pooling for database operations
- Flexible configuration management

## Prerequisites

- Kubernetes cluster
- MySQL/MariaDB database
- Go 1.21 or later
- Docker
- kubectl

## Installation

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/your-org/operator-uuwaf.git
cd operator-uuwaf
```

2. Install dependencies:
```bash
go mod download
```

3. Build the operator:
```bash
go build -o operator-uuwaf
```

### Building with Docker

1. Build the Docker image:
```bash
docker build -t your-registry/uuwaf-operator:latest .
```

2. Push the image to your registry:
```bash
docker push your-registry/uuwaf-operator:latest
```

## Configuration

The operator can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| MYSQL_HOST | MySQL host address | localhost |
| MYSQL_PORT | MySQL port | 3306 |
| MYSQL_USER | MySQL username | root |
| MYSQL_PASSWORD | MySQL password | password |
| MYSQL_DATABASE | MySQL database name | uuwaf |
| METRICS_PORT | Port for Prometheus metrics | 8080 |

## Deployment

### Kubernetes Deployment

1. Create a namespace (optional):
```bash
kubectl create namespace uuwaf
```

2. Create the database secret:
```bash
kubectl create secret generic uuwaf-db-secret \
  --from-literal=MYSQL_USER=root \
  --from-literal=MYSQL_PASSWORD=your-password \
  -n uuwaf
```

3. Deploy the operator:
```bash
kubectl apply -f deploy/kubernetes/deployment.yaml
```

4. Verify the deployment:
```bash
kubectl get pods -n uuwaf
kubectl logs -f deployment/uuwaf-operator -n uuwaf
```

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Run the operator:
```bash
go run main.go
```

## Monitoring

### Prometheus Metrics

The operator exposes Prometheus metrics on port 8080. You can access them at:
```
http://localhost:8080/metrics
```

Available metrics:
- `uuwaf_audit_events_total`: Total number of audit events processed
- `uuwaf_pod_reloads_total`: Total number of pod reloads
- `uuwaf_pod_reload_errors_total`: Total number of pod reload errors
- `uuwaf_db_connection_errors_total`: Total number of database connection errors

### Logging

The operator uses structured logging with zap. Example log output:
```json
{"level":"info","ts":1647123456.789,"msg":"Starting UUWAF operator","version":"1.0.0"}
{"level":"info","ts":1647123456.790,"msg":"Connected to database","host":"localhost","port":3306}
```

## Development

### Prerequisites

- Go 1.21 or later
- Docker
- kubectl
- make (optional)

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Run tests:
```bash
go test ./...
```

3. Build locally:
```bash
go build -o operator-uuwaf
```

### Docker Development

1. Build the development image:
```bash
docker build -t uuwaf-operator:dev .
```

2. Run the container:
```bash
docker run -it --rm \
  -e MYSQL_HOST=host.docker.internal \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=password \
  uuwaf-operator:dev
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request


## Authors

- Nguyen Anh Tuan (tuan.gli@gmail.com)

## Acknowledgments

- Thanks to all contributors
- Inspired by the Kubernetes Operator pattern 