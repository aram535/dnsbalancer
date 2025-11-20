# Quick Start Guide

Get **dnsbalancer** up and running in 5 minutes!

## Prerequisites

- Go 1.21+ installed
- Root/sudo access (for port 53)
- Two DNS servers to load balance (e.g., `192.168.1.2:53` and `192.168.1.3:53`)

## Step 1: Build

```bash
cd dnsbalancer
./build.sh
```

Or using make:
```bash
make build
```

## Step 2: Generate Configuration

```bash
./dnsbalancer genconfig
```

This creates `config.yaml` in the current directory.

## Step 3: Edit Configuration

Edit `config.yaml` and update the backend addresses to match your DNS servers:

```yaml
backends:
  - address: "192.168.1.2:53"   # Replace with your DNS server 1
  - address: "192.168.1.3:53"   # Replace with your DNS server 2
```

## Step 4: Validate

```bash
./dnsbalancer validate
```

Expected output:
```
✅ Configuration is VALID
```

## Step 5: Test Backends

```bash
./dnsbalancer healthcheck
```

Expected output:
```
[1/2] Testing 192.168.1.2:53 ... ✅ OK (15ms)
[2/2] Testing 192.168.1.3:53 ... ✅ OK (12ms)

✅ All backends are healthy
```

## Step 6: Run (Debug Mode)

Start the load balancer in debug mode to see it working:

```bash
sudo ./dnsbalancer serve --debug
```

You should see:
```
INFO[...] Starting dnsbalancer version=1.0.0
INFO[...] Registered backend backend=192.168.1.2:53
INFO[...] Registered backend backend=192.168.1.3:53
INFO[...] Health checking enabled
INFO[...] DNS load balancer started address=0.0.0.0:53
```

## Step 7: Test It!

In another terminal, send some test queries:

```bash
# Test basic query
dig @127.0.0.1 google.com

# Test multiple queries
for i in {1..10}; do
  dig @127.0.0.1 example.com +short
done
```

Watch the debug output to see queries being distributed across backends!

## Step 8: Production Deployment

### Option A: Run Directly

Stop the debug instance (Ctrl+C) and run in production mode:

```bash
sudo ./dnsbalancer serve
```

Logs will be written to `/var/log/dnsbalancer/dnsbalancer.log`

### Option B: Install as Systemd Service

```bash
# Install binary and service
sudo make install
sudo cp dnsbalancer.service /etc/systemd/system/
sudo systemctl daemon-reload

# Generate system-wide config
sudo mkdir -p /etc/dnsbalancer
sudo ./dnsbalancer genconfig -o /etc/dnsbalancer/config.yaml

# Edit config
sudo nano /etc/dnsbalancer/config.yaml

# Start service
sudo systemctl enable --now dnsbalancer

# Check status
sudo systemctl status dnsbalancer
```

### Option C: Docker

```bash
# Build and run with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f

# Test
dig @127.0.0.1 google.com
```

## Common Issues

### "Permission denied" on port 53

Port 53 requires root privileges. Either:

```bash
# Run with sudo
sudo ./dnsbalancer serve

# OR use setcap
sudo setcap 'cap_net_bind_service=+ep' ./dnsbalancer
./dnsbalancer serve
```

### "Port already in use"

Another service (like systemd-resolved) might be using port 53:

```bash
# Check what's using the port
sudo lsof -i :53

# If it's systemd-resolved
sudo systemctl stop systemd-resolved
```

### "Backend marked unhealthy"

Check if your backend DNS server is actually running and reachable:

```bash
# Test backend directly
dig @192.168.1.2 google.com

# Check network connectivity
nc -u -v 192.168.1.2 53
```

## Next Steps

- **Enable Health Checking**: Edit `config.yaml` and set `health_check.enabled: true`
- **Adjust Timeouts**: Tune `timeout` and health check intervals for your network
- **Monitor Logs**: `tail -f /var/log/dnsbalancer/dnsbalancer.log`
- **Read Full Docs**: See [README.md](README.md) for all features
- **Testing Guide**: See [TESTING.md](TESTING.md) for comprehensive testing scenarios

## Configuration Tips

### Minimal Configuration

```yaml
listen: "0.0.0.0:53"
backends:
  - address: "192.168.1.2:53"
  - address: "192.168.1.3:53"
```

### Recommended Production Configuration

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

## Getting Help

- **Validate Config**: `./dnsbalancer validate`
- **Test Backends**: `./dnsbalancer healthcheck`
- **Debug Mode**: Add `--debug` flag to see detailed logs
- **View Logs**: `tail -f /var/log/dnsbalancer/dnsbalancer.log`
- **Check Version**: `./dnsbalancer version`

## Success!

You now have a working DNS load balancer! 🎉

Your clients can now use your load balancer IP as their DNS server, and queries will be automatically distributed across your backends with health checking.

To make this your primary DNS server:
1. Update your DHCP server to advertise this IP as the DNS server
2. Or manually configure clients to use this IP for DNS
3. Watch the magic happen in the logs!

For more advanced configuration and deployment options, see the [README.md](README.md).
