# UUWAF Operator

**Author:** Nguyen Anh Tuan  
**Email:** tuan.gli@gmail.com

The UUWAF Operator is a service that monitors the UUWAF audit database and automatically reloads UUWAF configurations when changes are detected in the Site or Rule configurations.

## Prerequisites

- Go 1.16 or higher
- MySQL/MariaDB
- Kubernetes cluster access
- Root access on the Linux system where the operator will run

## Building the Operator

### Building on Linux
1. Clone the repository:
```bash
git clone <repository-url>
cd operator-uuwaf
```

2. Build the binary:
```bash
go build -o uuwaf-operator main.go
```

### Building on macOS for Linux Deployment
If you're building on macOS for deployment on Linux, you need to cross-compile for Linux:

1. Clone the repository:
```bash
git clone <repository-url>
cd operator-uuwaf
```

2. Set environment variables for cross-compilation:
```bash
export GOOS=linux
export GOARCH=amd64
```

3. Build the binary:
```bash
go build -o uuwaf-operator main.go
```

The resulting binary will be compatible with Linux systems even though it was built on macOS.

## Installation and Setup

1. Create the operator directory:
```bash
sudo mkdir -p /opt/uuwaf-operator
```

2. Copy the binary to the operator directory:
```bash
sudo cp uuwaf-operator /opt/uuwaf-operator/
```

3. Create the environment configuration file:
```bash
sudo mkdir -p /etc/sysconfig
sudo cp uuwaf-operator /etc/sysconfig/uuwaf-operator
sudo chmod 600 /etc/sysconfig/uuwaf-operator
```

4. Create and enable the systemd service:
```bash
sudo cp uuwaf-operator.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable uuwaf-operator
sudo systemctl start uuwaf-operator
```

## Configuration

The operator can be configured through environment variables in `/etc/sysconfig/uuwaf-operator`:

```ini
MYSQL_USER=root
MYSQL_PASSWORD=<password>
MYSQL_HOST=<ip or domain>
MYSQL_PORT=3306
MYSQL_DATABASE=uuwaf
POD_NAMESPACE=uusec
```

## Service Management

- Check service status:
```bash
sudo systemctl status uuwaf-operator
```

- View logs:
```bash
sudo journalctl -u uuwaf-operator -f
```

- Restart service:
```bash
sudo systemctl restart uuwaf-operator
```

- Stop service:
```bash
sudo systemctl stop uuwaf-operator
```

## How It Works

1. The operator connects to the MySQL database and monitors the `waf_audits` table
2. When a new audit record is detected with type "Site" or "Rule":
   - It identifies all UUWAF pods in the specified namespace
   - Executes the `uuwaf -s reload` command on each pod
   - Logs the results of the reload operation

## Troubleshooting

1. Check service status:
```bash
sudo systemctl status uuwaf-operator
```

2. View detailed logs:
```bash
sudo journalctl -u uuwaf-operator -f
```

3. Verify environment variables:
```bash
sudo cat /etc/sysconfig/uuwaf-operator
```

4. Check database connectivity:
```bash
mysql -h <MYSQL_HOST> -u <MYSQL_USER> -p<MYSQL_PASSWORD> <MYSQL_DATABASE>
```

## Security Considerations

- The environment file `/etc/sysconfig/uuwaf-operator` contains sensitive information and should be protected with appropriate permissions (600)
- The service runs as root to ensure it has necessary permissions to execute commands on UUWAF pods
- Consider implementing additional security measures based on your organization's requirements

## Support

For issues and support, please contact the development team or create an issue in the repository. 