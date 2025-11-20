# Testing Guide for dnsbalancer

This guide covers testing and validating dnsbalancer in different scenarios.

## Prerequisites

- Go 1.21 or later installed
- Root/sudo access (for binding to port 53)
- Two or more DNS servers to use as backends
- `dig` or `nslookup` for testing queries

## Building

```bash
# Option 1: Use the build script
./build.sh

# Option 2: Use make
make build

# Option 3: Manual build
go build -o dnsbalancer .
```

## Test Scenarios

### 1. Configuration Testing

#### Generate and Validate Config

```bash
# Generate example config
./dnsbalancer genconfig

# Edit config.yaml with your backends
nano config.yaml

# Validate configuration
./dnsbalancer validate
```

#### Expected Output
```
✅ Configuration is VALID

Summary:
  Listen Address:    0.0.0.0:53
  Timeout:           3s
  ...
```

### 2. Backend Connectivity Testing

Test that your backends are reachable before starting the load balancer:

```bash
./dnsbalancer healthcheck
```

#### Expected Output
```
Testing 2 backends with query: . (NS)
Timeout: 3s

[1/2] Testing 192.168.1.2:53 ... ✅ OK (15ms)
[2/2] Testing 192.168.1.3:53 ... ✅ OK (12ms)

✅ All backends are healthy
```

#### Testing Specific Queries

```bash
# Test with specific query
./dnsbalancer healthcheck --query example.com --type A

# Test with longer timeout
./dnsbalancer healthcheck --timeout 5s
```

### 3. Load Balancer Functionality

#### Start in Debug Mode

```bash
sudo ./dnsbalancer serve --debug
```

This will:
- Log to console instead of file
- Show detailed query forwarding information
- Display health check results

#### Send Test Queries

In another terminal:

```bash
# Test basic query
dig @127.0.0.1 google.com

# Test multiple queries to verify round-robin
for i in {1..10}; do
  dig @127.0.0.1 example.com +short
done

# Test with nslookup
nslookup google.com 127.0.0.1
```

#### Verify Round-Robin

Watch the debug logs to see queries being distributed:

```
INFO[...] Forwarding query to backend backend=192.168.1.2:53 client=127.0.0.1:xxxxx
INFO[...] Forwarding query to backend backend=192.168.1.3:53 client=127.0.0.1:xxxxx
INFO[...] Forwarding query to backend backend=192.168.1.2:53 client=127.0.0.1:xxxxx
```

### 4. Health Check Testing

#### Test Automatic Failover

1. Start dnsbalancer with health checking enabled:

```yaml
health_check:
  enabled: true
  interval: 5s
  failure_threshold: 2
```

```bash
sudo ./dnsbalancer serve --debug
```

2. Stop one backend DNS server

3. Watch logs for health check failures:

```
WARN[...] Health check failed backend=192.168.1.2:53
WARN[...] Backend marked unhealthy backend=192.168.1.2:53 consecutive_fails=2
```

4. Send queries - they should only go to healthy backend:

```bash
for i in {1..5}; do
  dig @127.0.0.1 google.com +short
done
```

5. Start the backend again and watch it recover:

```
INFO[...] Backend recovered and marked healthy backend=192.168.1.2:53
```

### 5. Fail Behavior Testing

#### Test Fail-Closed Behavior

Config:
```yaml
fail_behavior: closed
```

Steps:
1. Stop ALL backend DNS servers
2. Send queries
3. Queries should be dropped (no response)

```bash
dig @127.0.0.1 google.com
# Should timeout with no response
```

#### Test Fail-Open Behavior

Config:
```yaml
fail_behavior: open
```

Steps:
1. Stop ALL backend DNS servers
2. Send queries
3. Queries should still attempt to forward (will fail but tries)

### 6. Performance Testing

#### Test Query Throughput

```bash
# Install dnsperf (if not already installed)
# Ubuntu/Debian: apt-get install dnsperf

# Create test query file
cat > queries.txt << EOF
google.com A
example.com A
cloudflare.com A
EOF

# Run performance test
dnsperf -s 127.0.0.1 -d queries.txt -c 10 -l 30
```

This will:
- Send queries from 10 concurrent clients
- Run for 30 seconds
- Report queries per second and latency

#### Expected Performance
- 5,000-10,000+ queries/second on modern hardware
- Sub-millisecond overhead added by load balancer

### 7. Logging Tests

#### File Logging

Normal mode (logs to file):

```bash
sudo ./dnsbalancer serve

# In another terminal, watch logs
tail -f /var/log/dnsbalancer/dnsbalancer.log
```

#### Log Levels

Test different log levels:

```bash
# Info level (default)
sudo ./dnsbalancer serve --log-level info

# Debug level (verbose)
sudo ./dnsbalancer serve --log-level debug

# Warn level (errors/warnings only)
sudo ./dnsbalancer serve --log-level warn
```

### 8. Edge Cases

#### Test Empty Backend Pool

Config with no backends:
```yaml
backends: []
```

```bash
./dnsbalancer validate
# Should fail validation
```

#### Test Invalid Backend Address

Config:
```yaml
backends:
  - address: "not-a-valid-address"
```

```bash
./dnsbalancer healthcheck
# Should show connection failure
```

#### Test Timeout Handling

Config:
```yaml
timeout: 100ms  # Very short timeout
```

Send queries and verify timeouts are handled gracefully.

### 9. Systemd Integration

#### Test Systemd Service

```bash
# Install service
sudo cp dnsbalancer.service /etc/systemd/system/
sudo systemctl daemon-reload

# Start service
sudo systemctl start dnsbalancer

# Check status
sudo systemctl status dnsbalancer

# View logs
sudo journalctl -u dnsbalancer -f

# Test queries
dig @127.0.0.1 google.com

# Stop service
sudo systemctl stop dnsbalancer
```

### 10. Docker Testing

#### Build and Run in Docker

```bash
# Build image
docker build -t dnsbalancer:test .

# Run container
docker run -d \
  --name dnsbalancer-test \
  --network host \
  -v $(pwd)/config.yaml:/etc/dnsbalancer/config.yaml:ro \
  dnsbalancer:test

# Check logs
docker logs -f dnsbalancer-test

# Test queries
dig @127.0.0.1 google.com

# Stop container
docker stop dnsbalancer-test
docker rm dnsbalancer-test
```

#### Docker Compose Testing

```bash
# Start with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f

# Test queries
dig @127.0.0.1 google.com

# Stop
docker-compose down
```

## Troubleshooting Tests

### Port Already in Use

```bash
# Check what's using port 53
sudo lsof -i :53
sudo netstat -tulpn | grep :53

# Stop conflicting service
sudo systemctl stop systemd-resolved  # Common on Ubuntu
```

### Permission Issues

```bash
# Verify you're running with sudo
sudo ./dnsbalancer serve

# Or use setcap
sudo setcap 'cap_net_bind_service=+ep' ./dnsbalancer
./dnsbalancer serve
```

### Backend Connection Issues

```bash
# Test backend directly
dig @192.168.1.2 google.com

# Check network connectivity
nc -u -v 192.168.1.2 53

# Verify firewall rules
sudo iptables -L -n | grep 53
```

## Automated Testing Script

Create a test script:

```bash
#!/bin/bash
# test-dnsbalancer.sh

set -e

echo "=== dnsbalancer Test Suite ==="

echo -e "\n1. Building..."
./build.sh

echo -e "\n2. Generating config..."
./dnsbalancer genconfig -o test-config.yaml

echo -e "\n3. Validating config..."
./dnsbalancer validate --config test-config.yaml

echo -e "\n4. Testing backends..."
./dnsbalancer healthcheck --config test-config.yaml

echo -e "\n5. Starting server (will run for 10 seconds)..."
sudo timeout 10s ./dnsbalancer serve --config test-config.yaml --debug &
SERVER_PID=$!

sleep 2

echo -e "\n6. Sending test queries..."
for i in {1..5}; do
    dig @127.0.0.1 google.com +short > /dev/null && echo "Query $i: OK" || echo "Query $i: FAILED"
done

echo -e "\n7. Stopping server..."
sudo kill $SERVER_PID 2>/dev/null || true

echo -e "\n✅ All tests passed!"
```

Make it executable:
```bash
chmod +x test-dnsbalancer.sh
./test-dnsbalancer.sh
```

## Success Criteria

A successful test run should demonstrate:

1. ✅ Configuration validates without errors
2. ✅ All backends pass health checks
3. ✅ Server starts and binds to port 53
4. ✅ Queries are successfully forwarded and responses returned
5. ✅ Round-robin distribution is working
6. ✅ Health checks detect and remove unhealthy backends
7. ✅ Health checks detect and restore recovered backends
8. ✅ Graceful shutdown works correctly
9. ✅ Logs are written to expected location
10. ✅ No memory leaks or goroutine leaks after extended operation
