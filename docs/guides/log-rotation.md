# Log Rotation Guide

This guide explains how lowkey manages log files through automatic rotation.

## Overview

Lowkey implements automatic log rotation to prevent log files from growing indefinitely. When the active log file reaches a size threshold, it is renamed with a timestamp and a new log file is created.

## Rotation Behavior

### When Rotation Occurs

Rotation is triggered when:
- The active log file (`lowkey.log`) reaches **10 MB** in size
- A new write would cause the file to exceed this threshold

**Important**: Rotation happens **before** the write that would exceed the limit, ensuring no single log file exceeds 10 MB.

### Archive Naming

When rotation occurs, the active log file is renamed with a timestamp:

```
lowkey.log → lowkey.log.20250115-143052
```

**Format**: `lowkey.log.YYYYMMDD-HHMMSS`

**Example Archive Names**:
- `lowkey.log.20250115-143052` (rotated at 2025-01-15 14:30:52)
- `lowkey.log.20250115-180421` (rotated at 2025-01-15 18:04:21)
- `lowkey.log.20250116-091230` (rotated at 2025-01-16 09:12:30)

### Archive Retention

Lowkey keeps a maximum of **5 archived log files**.

When a 6th archive would be created, the oldest archive is automatically deleted.

**Example Timeline**:
```
1. lowkey.log (active)
2. lowkey.log (reaches 10 MB)
   → lowkey.log.20250115-100000 (archive 1)
   → lowkey.log (new active file)

3. lowkey.log (reaches 10 MB again)
   → lowkey.log.20250115-110000 (archive 2)
   → lowkey.log (new active file)

... continues until ...

6. Total archives: 5
   - lowkey.log.20250115-100000 (oldest)
   - lowkey.log.20250115-110000
   - lowkey.log.20250115-120000
   - lowkey.log.20250115-130000
   - lowkey.log.20250115-140000 (newest)

7. Next rotation:
   → lowkey.log.20250115-100000 is DELETED
   → lowkey.log.20250115-150000 is CREATED
```

### Maximum Disk Usage

**Calculation**:
- Active file: up to 10 MB
- Archives: 5 × 10 MB = 50 MB
- **Total**: up to 60 MB

This is the maximum disk space lowkey will use for logs under normal operation.

## Viewing Logs

### Active Log

Use `lowkey tail` to follow the active log file:

```bash
# Follow active log in real-time
lowkey tail

# View last 50 lines
lowkey tail | tail -n 50
```

### Archived Logs

View archived logs directly from the state directory:

```bash
# Linux
ls ~/.local/state/lowkey/lowkey.log.*
cat ~/.local/state/lowkey/lowkey.log.20250115-143052

# macOS
ls ~/Library/Application\ Support/lowkey/lowkey.log.*
cat ~/Library/Application\ Support/lowkey/lowkey.log.20250115-143052

# Windows
dir %LOCALAPPDATA%\lowkey\lowkey.log.*
type %LOCALAPPDATA%\lowkey\lowkey.log.20250115-143052
```

### All Logs Combined

View active log plus all archives in chronological order:

```bash
# Linux/macOS
cat ~/.local/state/lowkey/lowkey.log.* ~/.local/state/lowkey/lowkey.log

# With timestamps sorted
ls -rt ~/.local/state/lowkey/lowkey.log* | xargs cat
```

### Search Across All Logs

```bash
# Search for errors in all log files
grep -i error ~/.local/state/lowkey/lowkey.log*

# Search with context (3 lines before and after)
grep -i -C 3 "panic" ~/.local/state/lowkey/lowkey.log*

# Count errors by archive
for log in ~/.local/state/lowkey/lowkey.log*; do
  echo "$log: $(grep -c "error" $log)"
done
```

## Manual Rotation

### Trigger Rotation

Rotation happens automatically, but you can trigger it manually by stopping and starting the daemon:

```bash
# Stop daemon (closes active log file)
lowkey stop

# Start daemon (opens new log file)
lowkey start /path/to/watch
```

**Note**: This creates a new log file but doesn't rename the old one with a timestamp. The old file remains as `lowkey.log` until the next automatic rotation.

### Force New Log File

To start with a fresh log file:

```bash
# Stop daemon
lowkey stop

# Rename or remove current log
mv ~/.local/state/lowkey/lowkey.log ~/.local/state/lowkey/lowkey.log.backup

# Start daemon (creates new log)
lowkey start /path/to/watch
```

## Rotation Process Details

### Step-by-Step Process

When rotation is triggered:

1. **Close active file**: The current `lowkey.log` is closed
2. **Rename with timestamp**: `lowkey.log` → `lowkey.log.YYYYMMDD-HHMMSS`
3. **Check archive count**: Count existing `lowkey.log.*` files
4. **Remove excess archives**: If more than 5, delete oldest ones
5. **Create new active file**: Open new `lowkey.log` for writing
6. **Continue logging**: Resume writing to new active file

### Concurrent Safety

The rotation process is **thread-safe**:
- Uses mutex locks during rotation
- No log messages are lost during rotation
- Writes are serialized to prevent corruption

### Atomic Operations

Rotation uses atomic file operations:
- File rename is atomic on most filesystems
- No partial files are created
- Rotation either completes fully or doesn't happen

## Managing Log Disk Space

### Check Current Usage

```bash
# Linux/macOS
du -sh ~/.local/state/lowkey/

# Detailed breakdown
du -h ~/.local/state/lowkey/lowkey.log*

# Windows
dir %LOCALAPPDATA%\lowkey
```

### Clear Old Logs

Use the `lowkey clear` command:

```bash
# Remove all log files (active + archives)
lowkey clear --logs --yes

# Or manually
rm ~/.local/state/lowkey/lowkey.log*
```

**Note**: The daemon will create a new `lowkey.log` when it next writes.

### Archive Management

Keep specific archives:

```bash
# Archive logs before clearing
mkdir ~/lowkey-log-archive
cp ~/.local/state/lowkey/lowkey.log* ~/lowkey-log-archive/

# Clear current logs
lowkey clear --logs --yes

# Restore specific archive if needed
cp ~/lowkey-log-archive/lowkey.log.20250115-143052 ~/restored-log.txt
```

## Custom Log Paths

### Specify Custom Log Path

When starting the daemon, logs go to the state directory by default. To use a custom path, edit the manifest:

```json
{
  "directories": ["/path/to/watch"],
  "log_path": "/custom/path/to/logs/lowkey.log"
}
```

Then start with the manifest:

```bash
lowkey start --manifest custom-manifest.json
```

**Note**: Rotation behavior is the same regardless of log path.

### Multiple Daemon Instances

Each daemon instance maintains separate logs:

```bash
# Instance 1 (uses default state directory)
lowkey start /path/to/project-a

# Instance 2 (uses custom state directory)
XDG_STATE_HOME=/tmp/lowkey-b lowkey start /path/to/project-b
```

Logs are stored in:
- Instance 1: `~/.local/state/lowkey/lowkey.log`
- Instance 2: `/tmp/lowkey-b/lowkey/lowkey.log`

## Troubleshooting

### Rotation not happening

**Symptom**: Log file exceeds 10 MB

**Possible causes**:
1. **Single log line exceeds threshold**: If a single write is >10 MB, it won't be split
2. **Permission issues**: Can't rename or create files
3. **Disk full**: Can't create new files

**Solutions**:
```bash
# Check file size
ls -lh ~/.local/state/lowkey/lowkey.log

# Check permissions
ls -l ~/.local/state/lowkey/

# Manually rotate
lowkey stop
mv ~/.local/state/lowkey/lowkey.log ~/.local/state/lowkey/lowkey.log.manual
lowkey start /path/to/watch
```

### Archives not being deleted

**Symptom**: More than 5 archived files exist

**Possible causes**:
- Manual log files don't match the `lowkey.log.*` pattern
- Permission issues preventing deletion

**Solutions**:
```bash
# List all log files
ls -lh ~/.local/state/lowkey/

# Manually clean up
rm ~/.local/state/lowkey/lowkey.log.* # Removes all archives
```

### Can't find archived logs

**Symptom**: `lowkey.log.*` files don't exist after rotation

**Possible causes**:
- Rotation hasn't occurred yet (log file < 10 MB)
- Custom log path in use
- Logs were cleared

**Solutions**:
```bash
# Check current log size
ls -lh ~/.local/state/lowkey/lowkey.log

# Check manifest for custom log path
cat ~/.local/state/lowkey/daemon.json

# Verify rotation is working
# Force some log activity and check if rotation occurs
lowkey tail
```

### Log files taking too much space

**Symptom**: Log directory exceeds expected 60 MB

**Possible causes**:
- High log verbosity (trace mode enabled)
- Very frequent events

**Solutions**:

1. **Disable trace mode** (if not needed):
   ```bash
   lowkey stop
   lowkey start /path/to/watch  # Without --trace
   ```

2. **Clear logs more frequently**:
   ```bash
   # Add to cron or scheduled task
   lowkey clear --logs --yes
   ```

3. **Reduce event volume**:
   ```bash
   # Add more ignore patterns
   echo "**/*.log" >> .lowkey
   echo "**/node_modules/" >> .lowkey
   ```

## Performance Implications

### Rotation Overhead

Rotation is fast but not instant:
- **File rename**: ~1-5ms
- **Archive cleanup**: ~5-10ms per file
- **New file creation**: ~1-5ms

**Total rotation time**: ~10-50ms depending on filesystem

**Impact**: Negligible for most workloads. One log message may be delayed during rotation.

### Write Performance

Rotation doesn't impact steady-state write performance:
- Writes are buffered by the OS
- No locks held during normal writes
- Only rotation checks file size (fast operation)

## Related Documentation

- [Telemetry Guide](telemetry.md) - Logging and tracing configuration
- [State Directories](state-directories.md) - Where log files are stored
- [Troubleshooting](troubleshooting.md) - Common log-related issues
