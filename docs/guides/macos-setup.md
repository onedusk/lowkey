# macOS Setup Guide

This guide covers installation and daemon setup for lowkey on macOS.

## Installation

Build and install the lowkey binary:

```bash
# Build the binary
make build

# Copy to PATH
sudo cp ./lowkey/lowkey /usr/local/bin/
```

## Configure launchd Service

Create a launchd service to manage lowkey automatically.

Create `/Library/LaunchDaemons/dev.lowkey.daemon.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.lowkey.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/lowkey</string>
        <string>start</string>
        <string>--metrics</string>
        <string>127.0.0.1:9600</string>
        <string>--trace</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    <key>StandardOutPath</key>
    <string>/var/log/lowkey-daemon.out</string>
    <key>StandardErrorPath</key>
    <string>/var/log/lowkey-daemon.err</string>
</dict>
</plist>
```

## Start and Verify

Load and start the service:

```bash
# Load the daemon
sudo launchctl load /Library/LaunchDaemons/dev.lowkey.daemon.plist
sudo launchctl start dev.lowkey.daemon

# Verify status
lowkey status

# Tail logs
lowkey tail
```

## Managing the Service

```bash
# Stop the daemon
sudo launchctl stop dev.lowkey.daemon

# Unload the daemon
sudo launchctl unload /Library/LaunchDaemons/dev.lowkey.daemon.plist

# Restart the daemon
sudo launchctl stop dev.lowkey.daemon
sudo launchctl start dev.lowkey.daemon
```
