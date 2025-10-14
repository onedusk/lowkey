# State Directory Locations

This guide documents where lowkey stores its state files, including daemon manifests, logs, and cache data.

## Overview

Lowkey follows platform conventions for state file locations, adhering to the XDG Base Directory Specification on Linux and using native directories on macOS and Windows.

## Default Locations

### Linux

```bash
~/.local/state/lowkey/
```

**Example**: `/home/alice/.local/state/lowkey/`

### macOS

```bash
~/Library/Application Support/lowkey/
```

**Example**: `/Users/alice/Library/Application Support/lowkey/`

### Windows

```bash
%LOCALAPPDATA%\lowkey\
```

**Example**: `C:\Users\Alice\AppData\Local\lowkey\`

## Files Stored

The state directory contains the following files:

### daemon.json
- **Purpose**: Daemon manifest with configuration
- **Format**: JSON
- **Contents**: Watched directories, log path, ignore file location
- **Created**: When `lowkey start` is run
- **Deleted**: When `lowkey stop` is run (optional with `lowkey clear --state`)

### lowkey.log
- **Purpose**: Daemon log file
- **Format**: Plain text with timestamps
- **Rotation**: At 10 MB, keeps 5 archives
- **Archives**: `lowkey.log.1`, `lowkey.log.2`, etc.
- **Created**: When daemon starts
- **Viewed**: Via `lowkey tail` command

### lowkey.pid
- **Purpose**: Process ID file for running daemon
- **Format**: Plain text with single PID number
- **Created**: When daemon starts
- **Deleted**: When daemon stops cleanly
- **Usage**: Used by `lowkey stop` and `lowkey status`

### cache.json (optional)
- **Purpose**: File signature cache for incremental scanning
- **Format**: JSON with file paths and modification times
- **Created**: During daemon operation
- **Deleted**: Via `lowkey clear --state` or when daemon stops

## Finding Your State Directory

### Command-Line Discovery

```bash
# View daemon status (shows state directory path)
lowkey status

# Check environment variables
echo $XDG_STATE_HOME  # Linux (if set)
echo $LOCALAPPDATA    # Windows (if set)

# macOS - always uses Library/Application Support
ls ~/Library/Application\ Support/lowkey/
```

### Programmatic Discovery

The state directory resolution follows this priority:

1. **XDG_STATE_HOME** (if set): `$XDG_STATE_HOME/lowkey`
2. **Platform default**:
   - Linux: `~/.local/state/lowkey`
   - macOS: `~/Library/Application Support/lowkey`
   - Windows: `%LOCALAPPDATA%\lowkey`

## Environment Variable Overrides

### XDG_STATE_HOME (Linux/macOS)

Override the default state directory:

```bash
# Set custom state directory
export XDG_STATE_HOME=/custom/state/path

# Start daemon (uses /custom/state/path/lowkey/)
lowkey start /path/to/watch

# Verify location
lowkey status
```

### LOCALAPPDATA (Windows)

Override on Windows (usually set by system):

```powershell
# Check current value
echo $env:LOCALAPPDATA

# Set custom value (rare)
$env:LOCALAPPDATA = "C:\CustomAppData"

# Start daemon
lowkey start C:\path\to\watch
```

## Custom State Directories

### Per-Instance State

Run multiple daemon instances with separate state directories:

```bash
# Instance 1
XDG_STATE_HOME=/tmp/lowkey-instance-1 lowkey start /path/to/project-a

# Instance 2
XDG_STATE_HOME=/tmp/lowkey-instance-2 lowkey start /path/to/project-b

# Check status of each
XDG_STATE_HOME=/tmp/lowkey-instance-1 lowkey status
XDG_STATE_HOME=/tmp/lowkey-instance-2 lowkey status
```

### Temporary State

Use temporary directories for ephemeral monitoring:

```bash
# Create temp state directory
mkdir -p /tmp/lowkey-temp

# Run with temp state
XDG_STATE_HOME=/tmp/lowkey-temp lowkey start /path/to/watch

# Clean up when done
lowkey stop
rm -rf /tmp/lowkey-temp
```

## Viewing State Files

### List All State Files

```bash
# Linux
ls -lah ~/.local/state/lowkey/

# macOS
ls -lah ~/Library/Application\ Support/lowkey/

# Windows
dir %LOCALAPPDATA%\lowkey
```

### View Manifest

```bash
# Linux/macOS
cat ~/.local/state/lowkey/daemon.json

# Windows
type %LOCALAPPDATA%\lowkey\daemon.json

# Or use lowkey status
lowkey status
```

### View Logs

```bash
# Use lowkey tail (recommended)
lowkey tail

# Or read directly
# Linux
cat ~/.local/state/lowkey/lowkey.log

# macOS
cat ~/Library/Application\ Support/lowkey/lowkey.log

# Windows
type %LOCALAPPDATA%\lowkey\lowkey.log
```

## Cleaning State Files

### Remove All State

```bash
# Stop daemon and clear all state
lowkey stop
lowkey clear --state --yes
```

### Remove Logs Only

```bash
# Keep manifest, remove logs
lowkey clear --logs --yes
```

### Manual Cleanup

```bash
# Linux
rm -rf ~/.local/state/lowkey/

# macOS
rm -rf ~/Library/Application\ Support/lowkey/

# Windows (PowerShell)
Remove-Item -Recurse $env:LOCALAPPDATA\lowkey
```

## Permissions

### Required Permissions

The state directory requires:
- **Read**: To load manifests and logs
- **Write**: To create/update manifests, logs, PID files
- **Execute**: To list directory contents

### Permission Issues

If you encounter permission errors:

```bash
# Check permissions (Linux/macOS)
ls -ld ~/.local/state/lowkey/

# Fix permissions (Linux/macOS)
chmod 755 ~/.local/state/lowkey/

# Windows: Use File Explorer → Properties → Security
```

## State Directory Best Practices

### 1. Don't Manually Edit While Running

Avoid editing `daemon.json` while the daemon is running:

```bash
# ❌ Don't do this
vim ~/.local/state/lowkey/daemon.json  # While daemon is running

# ✅ Do this instead
lowkey stop
lowkey start /new/path/to/watch
```

### 2. Use lowkey Commands

Prefer lowkey commands over direct file manipulation:

```bash
# ✅ Use commands
lowkey tail           # View logs
lowkey status         # View manifest
lowkey clear --logs   # Clean logs

# ❌ Avoid direct access
cat ~/.local/state/lowkey/lowkey.log
rm ~/.local/state/lowkey/daemon.json
```

### 3. Backup Important State

If you rely on daemon configuration:

```bash
# Backup manifest
cp ~/.local/state/lowkey/daemon.json ~/backup/lowkey-manifest.json

# Restore if needed
cp ~/backup/lowkey-manifest.json ~/.local/state/lowkey/daemon.json
```

## Troubleshooting

### Can't find state directory

**Symptoms**: `lowkey status` shows "manifest not found"

**Solutions**:
```bash
# Verify daemon is running
ps aux | grep lowkey

# Check if daemon.json exists
# Linux
ls ~/.local/state/lowkey/daemon.json

# macOS
ls ~/Library/Application\ Support/lowkey/daemon.json

# Windows
dir %LOCALAPPDATA%\lowkey\daemon.json

# If missing, start daemon again
lowkey start /path/to/watch
```

### Permission denied

**Symptoms**: "permission denied" when starting daemon

**Solutions**:
```bash
# Check directory permissions
ls -ld ~/.local/state/lowkey/

# Create directory if missing
mkdir -p ~/.local/state/lowkey
chmod 755 ~/.local/state/lowkey

# Or use custom directory with write access
XDG_STATE_HOME=/tmp/lowkey-state lowkey start /path/to/watch
```

### State directory full

**Symptoms**: "no space left on device" or slow performance

**Solutions**:
```bash
# Check disk usage
du -sh ~/.local/state/lowkey/

# Clear old logs
lowkey clear --logs --yes

# Or move to larger partition
export XDG_STATE_HOME=/mnt/larger-disk/state
lowkey start /path/to/watch
```

### Stale PID file

**Symptoms**: "daemon already running" but `ps` shows no process

**Solutions**:
```bash
# Remove stale PID file
rm ~/.local/state/lowkey/lowkey.pid

# Or use clear command
lowkey clear --state --yes

# Restart daemon
lowkey start /path/to/watch
```

## Platform-Specific Notes

### Linux

- Follows XDG Base Directory Specification
- Respects `XDG_STATE_HOME` environment variable
- Default: `~/.local/state/lowkey/`

### macOS

- Uses native Application Support directory
- Does not use `XDG_STATE_HOME` by default
- Default: `~/Library/Application Support/lowkey/`
- Hidden in Finder (use Terminal or Go → Go to Folder)

### Windows

- Uses `LOCALAPPDATA` for per-user data
- Does not use `XDG_STATE_HOME`
- Default: `%LOCALAPPDATA%\lowkey\`
- Path example: `C:\Users\Username\AppData\Local\lowkey\`

## Related Documentation

- [Manifest Schema](manifest-schema.md) - Structure of daemon.json
- [Telemetry Guide](telemetry.md) - Metrics and logging configuration
- [Troubleshooting](../README.md#faq) - Common issues and solutions
