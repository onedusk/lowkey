# Manifest Schema

This guide documents the structure of the `daemon.json` manifest file used by lowkey's daemon mode.

## Overview

The manifest file stores the daemon's configuration and is persisted to the state directory when you run `lowkey start`. The daemon reads this manifest on startup to determine which directories to watch and how to configure logging and ignore patterns.

## File Location

The manifest is stored as `daemon.json` in the platform-specific state directory:

| Platform | Default Path |
|----------|-------------|
| **Linux** | `~/.local/state/lowkey/daemon.json` |
| **macOS** | `~/Library/Application Support/lowkey/daemon.json` |
| **Windows** | `%LOCALAPPDATA%\lowkey\daemon.json` |

You can override the state directory using the `XDG_STATE_HOME` environment variable:
```bash
export XDG_STATE_HOME=/custom/path
lowkey start /path/to/watch
# Manifest saved to: /custom/path/lowkey/daemon.json
```

## JSON Schema

### Root Object

```json
{
  "directories": ["string"],
  "log_path": "string",
  "ignore_file": "string"
}
```

### Field Definitions

#### `directories` (required)
- **Type**: Array of strings
- **Description**: List of absolute paths to directories that the daemon should monitor
- **Validation**:
  - Must contain at least one directory
  - All paths are normalized to absolute paths
  - Duplicate paths are removed
  - Paths are sorted lexicographically for consistency

#### `log_path` (optional)
- **Type**: String
- **Description**: Absolute path to the daemon's log file
- **Default**: `lowkey.log` in the state directory
- **Validation**: Normalized to absolute path if relative

#### `ignore_file` (optional)
- **Type**: String
- **Description**: Path to the `.lowkey` file containing ignore patterns
- **Default**: `.lowkey` file in each watched directory
- **Validation**: Normalized to absolute path if relative

## Example Manifests

### Minimal Manifest

```json
{
  "directories": [
    "/home/user/projects"
  ]
}
```

### Multi-Directory Manifest

```json
{
  "directories": [
    "/home/user/projects/backend",
    "/home/user/projects/frontend",
    "/var/www/html"
  ]
}
```

### Complete Manifest

```json
{
  "directories": [
    "/home/user/workspace/project-a",
    "/home/user/workspace/project-b"
  ],
  "log_path": "/var/log/lowkey/daemon.log",
  "ignore_file": "/home/user/.config/lowkey/ignore-patterns"
}
```

### Development Manifest

```json
{
  "directories": [
    "/Users/developer/code/web-app",
    "/Users/developer/code/api-service"
  ],
  "log_path": "/Users/developer/logs/lowkey.log"
}
```

## Path Normalization

Lowkey normalizes all paths in the manifest to ensure consistency:

1. **Absolute Path Conversion**: Relative paths are converted to absolute paths
2. **Cleaning**: Paths like `/foo/../bar` become `/bar`
3. **Deduplication**: Identical paths after normalization are removed
4. **Sorting**: Directories are sorted alphabetically

### Before Normalization
```json
{
  "directories": [
    ".",
    "/home/user/project/../project",
    "/home/user/project"
  ]
}
```

### After Normalization
```json
{
  "directories": [
    "/home/user/project"
  ]
}
```

## Creating Manifests

### Automatic Creation (Recommended)

The `lowkey start` command automatically creates and saves a manifest:

```bash
# Creates manifest with one directory
lowkey start /path/to/project

# Creates manifest with multiple directories
lowkey start /path/to/project-a /path/to/project-b
```

### Manual Creation

You can create a manifest file manually and load it:

```bash
# Create custom manifest
cat > my-manifest.json <<EOF
{
  "directories": [
    "/path/to/project"
  ],
  "log_path": "/custom/logs/lowkey.log"
}
EOF

# Start daemon with custom manifest
lowkey start --manifest my-manifest.json
```

## Viewing the Active Manifest

Use `lowkey status` to view the current manifest:

```bash
lowkey status
```

Example output:
```
Daemon: running
PID: 12345
Watched directories:
  - /home/user/projects/backend
  - /home/user/projects/frontend
Log path: /home/user/.local/state/lowkey/lowkey.log
```

## Modifying the Manifest

### Method 1: Stop and Restart

The safest way to modify the manifest is to stop the daemon and start it again:

```bash
# Stop daemon
lowkey stop

# Start with new directories
lowkey start /new/path/to/watch
```

### Method 2: Manual Edit + Reconciliation (Future)

Manual editing of the manifest file while the daemon is running is not currently supported but is planned for future releases:

```bash
# Edit manifest file
vim ~/.local/state/lowkey/daemon.json

# Trigger reconciliation (planned feature)
lowkey reconcile
```

## Validation Rules

### Required Fields

- `directories` must be present and contain at least one path

### Constraints

- **Empty directories**: Manifest will fail validation
  ```json
  {
    "directories": []  // ❌ Invalid
  }
  ```

- **Missing directories**: Manifest will fail validation
  ```json
  {
    "log_path": "/var/log/lowkey.log"  // ❌ Missing directories
  }
  ```

- **Null values**: Optional fields can be omitted or empty strings
  ```json
  {
    "directories": ["/path"],
    "log_path": "",      // ✓ Valid (uses default)
    "ignore_file": ""    // ✓ Valid (uses default)
  }
  ```

## Persistence Guarantees

Lowkey uses atomic writes to ensure manifest integrity:

1. **Atomic Writes**: Manifests are written to a temporary file and renamed atomically
2. **No Partial Writes**: Either the entire manifest is written or none of it is
3. **Corruption Protection**: Crashes during write operations don't corrupt the manifest

Implementation:
```
1. Create temp file: daemon-1234.json
2. Write manifest to temp file
3. Rename temp file to daemon.json (atomic operation)
```

## Error Handling

### Invalid JSON

If the manifest contains invalid JSON, the daemon will fail to start:

```bash
lowkey start /path/to/watch
# Error: config: decode manifest "/home/user/.local/state/lowkey/daemon.json": invalid character '}' ...
```

### Missing Directories

If watched directories don't exist, the daemon will still start but log warnings:

```json
{
  "directories": [
    "/nonexistent/path"
  ]
}
```

The daemon will monitor the parent path and detect when the directory is created.

## Advanced Usage

### Environment Variable Expansion

Lowkey does not perform environment variable expansion in manifests. Use absolute paths:

```json
{
  "directories": [
    "/home/user/project"  // ✓ Use absolute path
    // "$HOME/project"    // ❌ Not supported
  ]
}
```

To use environment variables, generate the manifest dynamically:

```bash
# Generate manifest with environment variables
cat > /tmp/manifest.json <<EOF
{
  "directories": [
    "$HOME/projects"
  ]
}
EOF

# Start with generated manifest
lowkey start --manifest /tmp/manifest.json
```

### Multiple Daemon Instances

Run multiple daemon instances by using different state directories:

```bash
# Instance 1: Monitor project A
XDG_STATE_HOME=/tmp/lowkey-a lowkey start /path/to/project-a

# Instance 2: Monitor project B
XDG_STATE_HOME=/tmp/lowkey-b lowkey start /path/to/project-b
```

Each instance maintains its own manifest in separate state directories.

## Troubleshooting

### Manifest not found

```bash
lowkey status
# Error: manifest not found
```

**Solution**: The daemon has not been started. Run `lowkey start <dirs>` first.

### Manifest corrupted

```bash
lowkey status
# Error: config: decode manifest: unexpected end of JSON input
```

**Solution**: Remove the corrupted manifest and restart:
```bash
rm ~/.local/state/lowkey/daemon.json
lowkey start /path/to/watch
```

### Permission denied

```bash
lowkey start /path/to/watch
# Error: state: create directory: permission denied
```

**Solution**: Ensure you have write permissions to the state directory or use a custom location:
```bash
XDG_STATE_HOME=/tmp/lowkey lowkey start /path/to/watch
```

## Related Documentation

- [State Directory Locations](state-directories.md) - Platform-specific state paths
- [Configuration & State](../README.md#configuration--state) - Overview of configuration options
- [Telemetry Guide](telemetry.md) - Metrics and tracing configuration
