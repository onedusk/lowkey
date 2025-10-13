# Linux Setup Guide

This guide covers installation and daemon setup for lowkey on Linux systems.

## Installation

Build and install the lowkey binary:

```bash
# Build the binary
make build

# Copy to PATH
sudo cp ./lowkey /usr/local/bin/
```

## Configure systemd Service

Create a systemd service to manage lowkey automatically.

Create `/etc/systemd/system/lowkey.service`:

```ini
[Unit]
Description=Lowkey Filesystem Monitor
After=network.target
Documentation=https://github.com/onedusk/lowkey

[Service]
Type=simple
ExecStart=/usr/local/bin/lowkey start --metrics 127.0.0.1:9600 --trace
Restart=on-failure
RestartSec=5s
StandardOutput=append:/var/log/lowkey-daemon.out
StandardError=append:/var/log/lowkey-daemon.err

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/var/log

[Install]
WantedBy=multi-user.target
```

## Start and Verify

Enable and start the service:

```bash
# Reload systemd to recognize new service
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable lowkey.service

# Start the service
sudo systemctl start lowkey.service

# Verify status
sudo systemctl status lowkey.service
lowkey status

# View logs
lowkey tail
# Or view systemd journal
sudo journalctl -u lowkey.service -f
```

## Managing the Service

```bash
# Stop the service
sudo systemctl stop lowkey.service

# Restart the service
sudo systemctl restart lowkey.service

# Disable auto-start on boot
sudo systemctl disable lowkey.service

# Check service status
sudo systemctl status lowkey.service

# View recent logs
sudo journalctl -u lowkey.service -n 50

# Follow logs in real-time
sudo journalctl -u lowkey.service -f
```

## Log Files

Lowkey creates log files in these locations:

- **Daemon stdout**: `/var/log/lowkey-daemon.out`
- **Daemon stderr**: `/var/log/lowkey-daemon.err`
- **Systemd journal**: `journalctl -u lowkey.service`

## Troubleshooting

### Service fails to start

Check the service status for error messages:
```bash
sudo systemctl status lowkey.service
sudo journalctl -u lowkey.service -n 50
```

### Permission denied errors

Ensure the binary has execute permissions:
```bash
sudo chmod +x /usr/local/bin/lowkey
```

Verify log directory permissions:
```bash
sudo touch /var/log/lowkey-daemon.out /var/log/lowkey-daemon.err
sudo chmod 644 /var/log/lowkey-daemon.out /var/log/lowkey-daemon.err
```

### Service not starting on boot

Verify the service is enabled:
```bash
sudo systemctl is-enabled lowkey.service
```

If disabled, enable it:
```bash
sudo systemctl enable lowkey.service
```
