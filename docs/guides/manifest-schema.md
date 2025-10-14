# Manifest JSON Schema

The `lowkey` daemon is configured via a JSON manifest file. This document outlines the structure and validation rules for the manifest.

## Overview

The manifest file is automatically created when you run `lowkey start` and is stored in the platform-specific state directory (see [State Directories](state-directories.md) for locations). The manifest persists the daemon's configuration and can be inspected or manually edited when the daemon is stopped.

## File Location

- **Linux**: `~/.local/state/lowkey/daemon.json`
- **macOS**: `~/Library/Application Support/lowkey/daemon.json`
- **Windows**: `%LOCALAPPDATA%\lowkey\daemon.json`

## Schema Structure

```json
{
  "directories": ["/path/to/watch", "/another/path"],
  "ignore_file": ".lowkey",
  "log_path": "",
  "metrics_address": "127.0.0.1:9600"
}
```

## Fields

### `directories` (required)

**Type:** Array of strings  
**Description:** List of absolute paths to directories that the daemon should monitor for filesystem events.

**Example:**
```json
"directories": [
  "/home/user/projects/webapp",
  "/home/user/projects/api"
]
```

**Validation Rules:**
- Must be an array with at least one element
- Each path must be an absolute path
- Paths must exist and be accessible
- Duplicate paths are allowed but inefficient

### `ignore_file` (optional)

**Type:** String  
**Description:** Name of the ignore patterns file (relative to each watched directory). Defaults to `.lowkey`.

**Example:**
```json
"ignore_file": ".lowkey"
```

**Default:** `.lowkey`

**Validation Rules:**
- Must be a non-empty string
- File is searched relative to each watched directory
- If the file doesn't exist, no patterns are ignored

### `log_path` (optional)

**Type:** String  
**Description:** Custom path for the daemon log file. If empty or omitted, logs are written to `lowkey.log` in the state directory.

**Example:**
```json
"log_path": "/var/log/lowkey/lowkey.log"
```

**Default:** `<state-directory>/lowkey.log`

**Validation Rules:**
- Must be an absolute path
- Parent directory must exist and be writable
- Log rotation applies to custom paths

### `metrics_address` (optional)

**Type:** String  
**Description:** Network address for the Prometheus metrics HTTP endpoint. Format: `host:port` or empty to disable metrics.

**Example:**
```json
"metrics_address": "127.0.0.1:9600"
```

**Default:** `` (metrics disabled)

**Validation Rules:**
- Must be a valid `host:port` format
- Port must be available (not in use)
- Use `127.0.0.1` to restrict to localhost
- Use `0.0.0.0` to expose on all interfaces (use with caution)

## Complete Example

```json
{
  "directories": [
    "/home/alice/workspace/backend",
    "/home/alice/workspace/frontend",
    "/home/alice/documents"
  ],
  "ignore_file": ".lowkey",
  "log_path": "/var/log/lowkey/custom.log",
  "metrics_address": "127.0.0.1:9600"
}
```

## Editing the Manifest

### Manual Editing

You can manually edit the manifest when the daemon is stopped:

```bash
# Stop the daemon
lowkey stop

# Edit the manifest
vim ~/.local/state/lowkey/daemon.json

# Start the daemon with new configuration
lowkey start
```

**Important:** Never edit the manifest while the daemon is running, as your changes may be overwritten.

### Programmatic Updates

The manifest is created automatically from command-line flags:

```bash
# This creates a manifest with the specified configuration
lowkey start --metrics 127.0.0.1:9600 --trace /path/to/monitor
```

## Validation Errors

Common validation errors and how to fix them:

### "directories field is required"

**Cause:** The `directories` array is empty or missing.  
**Fix:** Ensure at least one directory path is specified.

### "path must be absolute"

**Cause:** A relative path was provided in the `directories` array.  
**Fix:** Use absolute paths like `/home/user/project` instead of `./project`.

### "directory does not exist"

**Cause:** A path in the `directories` array doesn't exist.  
**Fix:** Create the directory or remove it from the manifest.

### "invalid metrics address format"

**Cause:** The `metrics_address` is not in valid `host:port` format.  
**Fix:** Use format like `127.0.0.1:9600` or leave empty to disable.

### "metrics port already in use"

**Cause:** Another process is using the specified metrics port.  
**Fix:** Choose a different port or stop the conflicting process.

## Advanced Usage

### Multiple Daemon Instances

Run multiple daemon instances with separate manifests using environment variables:

```bash
# Instance 1
lowkey start /project1

# Instance 2 (different state directory)
XDG_STATE_HOME=/tmp/lowkey-2 lowkey start /project2
```

Each instance maintains its own manifest in its state directory.

### Hot Reconfiguration

In future versions, `lowkey` will support hot reconfiguration by detecting manifest changes and reloading without restart. Currently, you must stop and restart the daemon for changes to take effect.

## Related Documentation

- [State Directories](state-directories.md) - Where manifests are stored
- [CLI Commands](../README.md#cli-commands) - Using `lowkey start` with flags
- [Troubleshooting](troubleshooting.md) - Common manifest-related issues
