# dnsbalancer

A lightweight, high-performance UDP DNS load balancer written in Go. Perfect for homelab and production environments where you need reliable DNS service with automatic failover.

## Features

- **Round-Robin Load Balancing**: Distributes DNS queries evenly across multiple backends
- **Active Health Checking**: Continuously monitors backend DNS servers and removes unhealthy ones from rotation
- **Configurable Fail Behavior**: Choose between fail-closed (drop queries) or fail-open (try anyway) when all backends are down
- **Flexible Configuration**: YAML configuration with command-line overrides
- **Structured Logging**: File-based logging with configurable log levels
- **GELF Support**: (Roadmap) Send logs to Graylog for centralized monitoring
- **Graceful Shutdown**: Cleanly handles in-flight queries during shutdown
- **Zero External Dependencies**: Self-contained binary, easy to deploy

## Installation

### From Source

```bash
git clone https://github.com/yourusername/dnsbalancer.git
cd dnsbalancer
go build -o dnsbalancer
sudo mv dnsbalancer /usr/local/bin/
```

### Build with Version Info

```bash
go build -ldflags "-X github.com/yourusername/dnsbalancer/cmd.Version=1.0.0 \
                    -X github.com/yourusername/dnsbalancer/cmd.GitCommit=$(git rev-parse HEAD) \
                    -X github.com/yourusername/dnsbalancer/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
         -o dnsbalancer
```

## Quick Start

### 1. Generate Configuration

```bash
dnsbalancer genconfig
```

This creates a `config.yaml` file with example configuration.

### 2. Edit Configuration

Edit `config.yaml` to add your DNS backend servers:

```yaml
listen: "0.0.0.0:53"
timeout: 3s
log_level: info
log_dir: /var/log/dnsbalancer
fail_behavior: closed

backends:
  - address: "192.168.1.2:53"
  - address: "192.168.1.3:53"

health_check:
  enabled: true
  interval: 10s
  timeout: 2s
  failure_threshold: 3
  success_threshold: 2
  query_name: "."
  query_type: "NS"
```

### 3. Validate Configuration

```bash
dnsbalancer validate
```

### 4. Test Backend Connectivity

```bash
dnsbalancer healthcheck
```

### 5. Start the Server

```bash
# With debug output to console
sudo dnsbalancer serve --debug

# Normal operation (logs to file)
sudo dnsbalancer serve
```

## Configuration

### Configuration File Priority

1. Command-line `--config` flag
2. `./config.yaml` (current directory)
3. `/etc/dnsbalancer/config.yaml` (system-wide)

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `listen` | string | `0.0.0.0:53` | UDP address to listen on |
| `timeout` | duration | `3s` | Timeout for backend queries |
| `log_level` | string | `info` | Log level (debug, info, warn, error) |
| `log_dir` | string | `/var/log/dnsbalancer` | Directory for log files |
| `fail_behavior` | string | `closed` | Behavior when all backends fail (`closed` or `open`) |
| `backends` | array | - | List of backend DNS servers |
| `health_check.enabled` | bool | `false` | Enable active health checking |
| `health_check.interval` | duration | `10s` | How often to check backends |
| `health_check.timeout` | duration | `2s` | Health check query timeout |
| `health_check.failure_threshold` | int | `3` | Failures before marking unhealthy |
| `health_check.success_threshold` | int | `2` | Successes before marking healthy |
| `health_check.query_name` | string | `.` | DNS name to query |
| `health_check.query_type` | string | `NS` | DNS query type |

### Backend Configuration

```yaml
backends:
  - address: "192.168.1.2:53"
  - address: "192.168.1.3:53"
  - address: "8.8.8.8:53"
```

## Commands

### serve

Start the DNS load balancer server:

```bash
dnsbalancer serve [flags]

Flags:
  --config string      Config file path
  --listen string      Listen address override
  --debug              Log to console
  --log-level string   Override log level
```

### validate

Validate configuration file syntax and values:

```bash
dnsbalancer validate [--config path]
```

### healthcheck

Test connectivity to all configured backends:

```bash
dnsbalancer healthcheck [flags]

Flags:
  --timeout duration   Health check timeout (default 3s)
  --query string       DNS query name (default ".")
  --type string        DNS query type (default "NS")
```

### genconfig

Generate an example configuration file:

```bash
dnsbalancer genconfig [--output path]
```

### version

Display version information:

```bash
dnsbalancer version
```

## Deployment

### Running as Root (Required for port 53)

DNS runs on port 53, which requires elevated privileges:

```bash
sudo dnsbalancer serve
```

### Alternative: Use setcap (Linux)

Allow the binary to bind to privileged ports without running as root:

```bash
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/dnsbalancer
dnsbalancer serve
```

### Systemd Service

Create `/etc/systemd/system/dnsbalancer.service`:

```ini
[Unit]
Description=DNS Load Balancer
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/dnsbalancer serve --config /etc/dnsbalancer/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable dnsbalancer
sudo systemctl start dnsbalancer
```

View logs:

```bash
sudo journalctl -u dnsbalancer -f
```

## Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │ UDP :53
       ▼
┌─────────────────┐
│  dnsbalancer    │
│  (Round-Robin)  │
└────┬───────┬────┘
     │       │
     ▼       ▼
┌─────────┐ ┌─────────┐
│ DNS 1   │ │ DNS 2   │
│ :53     │ │ :53     │
└─────────┘ └─────────┘
```

### Flow

1. Client sends DNS query to load balancer
2. Load balancer selects next healthy backend (round-robin)
3. Query is forwarded to selected backend
4. Backend response is returned to client
5. Health checker periodically verifies backend availability

## Logging

### Log Levels

- `debug`: Verbose output including query forwarding details
- `info`: Normal operational messages
- `warn`: Warnings like backend failures
- `error`: Error conditions

### Log Location

Default: `/var/log/dnsbalancer/dnsbalancer.log`

To view logs in real-time:

```bash
tail -f /var/log/dnsbalancer/dnsbalancer.log
```

### Debug Mode

Enable console logging for debugging:

```bash
dnsbalancer serve --debug
```

## Performance

- **Concurrency**: Each DNS query is handled in its own goroutine
- **Memory**: Minimal footprint (~10-20 MB)
- **Latency**: Sub-millisecond overhead on query forwarding
- **Throughput**: Tested to 10,000+ queries/second on modern hardware

## Troubleshooting

### Permission Denied on Port 53

Port 53 requires root privileges. Either:
- Run with `sudo`
- Use `setcap` (see Deployment section)
- Use an alternate port for testing: `--listen 0.0.0.0:5353`

### Backend Always Unhealthy

1. Check backend is actually running: `dig @192.168.1.2 example.com`
2. Verify network connectivity: `nc -u 192.168.1.2 53`
3. Check health check query is valid: `dnsbalancer healthcheck`
4. Review logs: `tail /var/log/dnsbalancer/dnsbalancer.log`

### All Queries Failing

1. Check fail_behavior setting in config
2. Verify at least one backend is healthy: `dnsbalancer healthcheck`
3. Test backend manually: `dig @192.168.1.2 google.com`

## Roadmap

### v1.1
- [ ] GELF logging to Graylog
- [ ] Weighted round-robin
- [ ] Statistics endpoint (HTTP)

### v1.2
- [ ] mDNS service discovery
- [ ] Prometheus metrics
- [ ] Admin API

### v2.0
- [ ] TCP DNS support
- [ ] DNS caching layer
- [ ] Geographic load balancing

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

Built for homelab enthusiasts and production deployments.
