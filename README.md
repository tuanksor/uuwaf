# UUWAF Operator

A Kubernetes operator that monitors UUWAF audit database and automatically reloads configurations when changes are detected.

## Features

- **Real-time Monitoring**: Monitors UUWAF audit database for configuration changes
- **Automatic Reload**: Automatically reloads UUWAF configurations when changes are detected
- **Structured Logging**: Uses Zap for structured logging with proper log levels
- **Metrics**: Exposes Prometheus metrics for monitoring
- **Graceful Shutdown**: Proper handling of SIGTERM and SIGINT signals
- **Connection Pooling**: Efficient database connection management
- **Configuration Management**: Flexible configuration through environment variables

## Prerequisites

- Kubernetes cluster
- MySQL/MariaDB database
- Go 1.21 or later

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/operator-uuwaf.git
cd operator-uuwaf
```

2. Install dependencies:
```bash
go mod download
```

3. Build the operator:
```bash
go build -o uuwaf-operator
```

## Configuration

The operator can be configured using environment variables:

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| MYSQL_USER | Database username | (required) |
| MYSQL_PASSWORD | Database password | (required) |
| MYSQL_HOST | Database host | (required) |
| MYSQL_PORT | Database port | 3306 |
| MYSQL_DATABASE | Database name | (required) |
| POD_NAMESPACE | Kubernetes namespace | uusec |

## Deployment

1. Create a Kubernetes secret for database credentials:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: uuwaf-db-secret
type: Opaque
data:
  MYSQL_USER: <base64-encoded-username>
  MYSQL_PASSWORD: <base64-encoded-password>
```

2. Deploy the operator:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: uuwaf-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: uuwaf-operator
  template:
    metadata:
      labels:
        app: uuwaf-operator
    spec:
      containers:
      - name: operator
        image: your-registry/uuwaf-operator:latest
        envFrom:
        - secretRef:
            name: uuwaf-db-secret
        env:
        - name: MYSQL_HOST
          value: "your-db-host"
        - name: MYSQL_DATABASE
          value: "uuwaf"
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
```

## Monitoring

The operator exposes Prometheus metrics at `/metrics` endpoint. Available metrics:

- `uuwaf_audit_events_processed_total`: Number of processed audit events
- `uuwaf_pod_reloads_total`: Number of pod reloads (success/failure)
- `uuwaf_processing_duration_seconds`: Time taken to process events
- `uuwaf_last_processed_id`: ID of the last processed audit event

## Logging

The operator uses structured logging with Zap. Log levels can be configured through the `LOG_LEVEL` environment variable.

Example log output:
```json
{"level":"info","ts":1234567890,"msg":"Starting monitoring","lastID":0}
{"level":"info","ts":1234567891,"msg":"New event found","id":123,"type":"Site","info":"site-config-changed"}
```

## Development

1. Install development dependencies:
```bash
go get -u github.com/prometheus/client_golang/prometheus
go get -u go.uber.org/zap
```

2. Run tests:
```bash
go test ./...
```

3. Build and run locally:
```bash
go run main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Authors

- Nguyen Anh Tuan (tuan.gli@gmail.com)

## Acknowledgments

- Kubernetes client-go library
- Prometheus client library
- Zap logging library 