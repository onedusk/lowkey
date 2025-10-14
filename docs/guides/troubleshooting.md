# Troubleshooting Guide

This guide helps diagnose and resolve common issues with lowkey daemon operation.

## Quick Diagnostics

Run these commands first to gather information:

```bash
# Check daemon status
lowkey status

# View recent logs
lowkey tail

# Check if process is running
ps aux | grep lowkey         # Linux/macOS
Get-Process lowkey          # Windows

# List state directory contents
ls -lah ~/.local/state/lowkey/                    # Linux
ls -lah ~/Library/Application\ Support/lowkey/    # macOS
dir %LOCALAPPDATA%\lowkey                         # Windows
```

## Daemon Won't Start

### "daemon already running"

**Symptom**: `lowkey start` returns "daemon already running with pid XXXX"

**Cause**: Another daemon instance is running, or stale PID file exists

**Solutions**:

1. Check if daemon is actually running:
   ```bash
   lowkey status
   ps aux | grep lowkey
   ```

2. Stop existing daemon:
   ```bash
   lowkey stop
   ```

3. If daemon won't stop, kill manually:
   ```bash
   # Find PID
   cat ~/.local/state/lowkey/lowkey.pid

   # Kill process
   kill <pid>          # Graceful
   kill -9 <pid>       # Force kill if needed
   ```

4. Remove stale PID file:
   ```bash
   lowkey clear --state --yes
   # Or manually
   rm ~/.local/state/lowkey/lowkey.pid
   ```

5. Start daemon again:
   ```bash
   lowkey start /path/to/watch
   ```

### "permission denied"

**Symptom**: `start: launch daemon: permission denied`

**Cause**: No write access to state directory or binary not executable

**Solutions**:

1. Check state directory permissions:
   ```bash
   ls -ld ~/.local/state/lowkey/
   ```

2. Create directory if missing:
   ```bash
   mkdir -p ~/.local/state/lowkey
   chmod 755 ~/.local/state/lowkey
   ```

3. Verify binary is executable:
   ```bash
   ls -l $(which lowkey)
   chmod +x $(which lowkey)
   ```

4. Use custom state directory:
   ```bash
   XDG_STATE_HOME=/tmp/lowkey-state lowkey start /path/to/watch
   ```

### "metrics: address already in use"

**Symptom**: Daemon fails to start with metrics endpoint error

**Cause**: Another process is using the metrics port

**Solutions**:

1. Check what's using the port:
   ```bash
   # Linux/macOS
   lsof -i :9600
   netstat -tulpn | grep 9600

   # Windows
   netstat -ano | findstr :9600
   ```

2. Use a different port:
   ```bash
   lowkey start --metrics 127.0.0.1:9601 /path/to/watch
   ```

3. Stop conflicting service:
   ```bash
   kill <pid-of-conflicting-process>
   ```

### "no such file or directory"

**Symptom**: `start: exec: no such file or directory`

**Cause**: Watched directory doesn't exist

**Solutions**:

1. Verify directory exists:
   ```bash
   ls -d /path/to/watch
   ```

2. Create directory if needed:
   ```bash
   mkdir -p /path/to/watch
   ```

3. Check for typos in path:
   ```bash
   # Use absolute paths
   lowkey start $(pwd)/relative/path
   ```

## Daemon Crashes or Restarts

### Unexpected restarts

**Symptom**: `lowkey status` shows high restart count

**Diagnosis**:

```bash
# Check logs for panic or errors
lowkey tail

# Look for patterns
grep -i "panic\|fatal\|error" ~/.local/state/lowkey/lowkey.log
```

**Common Causes**:

1. **Out of memory**: Watching too many files
   - Reduce watched directories
   - Add more ignore patterns
   - Increase system memory

2. **Filesystem permissions**: Can't access watched paths
   - Verify read permissions on all watched directories
   - Check for broken symlinks

3. **Disk full**: Can't write logs or cache
   - Check disk space: `df -h`
   - Clear old logs: `lowkey clear --logs --yes`

### Daemon stops unexpectedly

**Symptom**: `lowkey status` shows "not running" when it should be

**Solutions**:

1. Check system logs:
   ```bash
   # Linux
   journalctl -xe | grep lowkey
   dmesg | grep lowkey

   # macOS
   log show --predicate 'process == "lowkey"' --last 1h

   # Windows
   Get-EventLog -LogName Application -Source lowkey
   ```

2. Enable tracing for detailed logs:
   ```bash
   lowkey start --trace /path/to/watch
   lowkey tail
   ```

3. Check for OOM (Out of Memory) killer:
   ```bash
   # Linux
   dmesg | grep -i "killed process"
   grep -i "out of memory" /var/log/syslog
   ```

## Events Not Being Detected

### No events appearing

**Symptom**: File changes don't trigger events in `lowkey watch` or logs

**Diagnosis**:

1. Test with simple file:
   ```bash
   # Terminal 1
   lowkey watch /tmp/test

   # Terminal 2
   echo "test" > /tmp/test/file.txt
   ```

2. Check if files are ignored:
   ```bash
   # Review ignore patterns
   cat .lowkey

   # Test without ignore file
   mv .lowkey .lowkey.bak
   lowkey watch /path/to/watch
   ```

**Solutions**:

1. **Files are ignored**: Review `.lowkey` patterns
   ```bash
   # Temporarily disable ignores
   rm .lowkey
   ```

2. **Directory not watched**: Verify path in manifest
   ```bash
   lowkey status  # Check "Watched directories"
   ```

3. **Event buffer full**: Reduce event volume
   - Add more ignore patterns for busy directories
   - Watch fewer directories

4. **Polling fallback needed**: Force polling mode
   - Check if native events are broken
   - Test with small directory first

### Events delayed

**Symptom**: Changes take several seconds to appear

**Diagnosis**:

```bash
# Check event latency metrics
curl http://127.0.0.1:9600/metrics | grep latency
```

**Solutions**:

1. **Polling mode**: Events come every N seconds
   - This is expected behavior for polling
   - Reduce polling interval (if configurable)

2. **High system load**: CPU saturated
   ```bash
   top    # Check CPU usage
   ```
   - Reduce watched file count
   - Add ignore patterns for large directories

3. **Slow disk I/O**: Disk at capacity
   ```bash
   iostat -x 1 10  # Monitor disk performance
   ```
   - Clear disk space
   - Use faster storage

## Performance Issues

### High CPU usage

**Symptom**: `top` shows lowkey using >10% CPU consistently

**Diagnosis**:

```bash
# Check event rate
curl http://127.0.0.1:9600/metrics | grep events_total

# Monitor in real-time
watch -n 2 'curl -s http://127.0.0.1:9600/metrics | grep events'
```

**Solutions**:

1. **Too many events**: Reduce noise
   ```bash
   # Add ignore patterns for busy directories
   echo "node_modules/" >> .lowkey
   echo "**/*.log" >> .lowkey
   echo ".git/" >> .lowkey
   ```

2. **Watching too many files**: Narrow scope
   ```bash
   # Watch specific subdirectories instead of root
   lowkey start /project/src /project/config
   # Instead of
   lowkey start /project
   ```

3. **Insufficient ignore patterns**: Profile hotspots
   ```bash
   # Find directories with most changes
   lowkey watch /path | grep -o "path/to/.*" | sort | uniq -c | sort -rn
   ```

### High memory usage

**Symptom**: `lowkey` RSS exceeds 100 MB

**Diagnosis**:

```bash
# Check memory usage
ps aux | grep lowkey
top -p $(cat ~/.local/state/lowkey/lowkey.pid)
```

**Expected Memory**:
- Baseline: 15-25 MB
- Per watched directory: ~1-2 MB
- Per 10,000 ignore patterns: ~5 MB

**Solutions**:

1. **Watching too many directories**: Reduce scope
2. **Large ignore file**: Optimize patterns
3. **Cache size growth**: Clear cache
   ```bash
   lowkey clear --state --yes
   lowkey start /path/to/watch
   ```

## Log File Issues

### Logs not appearing

**Symptom**: `lowkey tail` shows no output or file not found

**Diagnosis**:

```bash
# Find log file location
lowkey status  # Shows log path

# Check if file exists
ls -lh ~/.local/state/lowkey/lowkey.log
```

**Solutions**:

1. **Daemon not started**: Start daemon first
   ```bash
   lowkey start /path/to/watch
   ```

2. **Custom log path**: Check manifest
   ```bash
   cat ~/.local/state/lowkey/daemon.json
   ```

3. **Permission error**: Fix permissions
   ```bash
   chmod 644 ~/.local/state/lowkey/lowkey.log
   ```

### Logs not rotating

**Symptom**: Single log file grows beyond 10 MB

**Expected Behavior**:
- Rotation at 10 MB
- 5 archives kept (`lowkey.log.1` through `lowkey.log.5`)

**Solutions**:

1. **Manual rotation**: Trigger by stopping/starting
   ```bash
   lowkey stop
   lowkey start /path/to/watch
   ```

2. **Disk full**: Check disk space
   ```bash
   df -h ~/.local/state/lowkey
   ```

3. **Permission issues**: Verify write access
   ```bash
   ls -lh ~/.local/state/lowkey/lowkey.log*
   ```

### "too many open files"

**Symptom**: Error in logs about file descriptor limit

**Diagnosis**:

```bash
# Check current limits
ulimit -n      # Soft limit
ulimit -Hn     # Hard limit

# Check process file descriptors
lsof -p $(cat ~/.local/state/lowkey/lowkey.pid) | wc -l
```

**Solutions**:

1. **Increase ulimit**:
   ```bash
   # Temporary
   ulimit -n 4096
   lowkey start /path/to/watch

   # Permanent (add to ~/.bashrc or ~/.zshrc)
   echo "ulimit -n 4096" >> ~/.bashrc
   ```

2. **Reduce watched files**: Add ignore patterns

3. **System-wide limit** (Linux):
   ```bash
   # Check system limit
   cat /proc/sys/fs/file-max

   # Increase if needed (requires root)
   echo "fs.file-max = 100000" | sudo tee -a /etc/sysctl.conf
   sudo sysctl -p
   ```

## State File Issues

### Corrupted manifest

**Symptom**: `config: decode manifest: unexpected end of JSON input`

**Solutions**:

1. **Remove and recreate**:
   ```bash
   rm ~/.local/state/lowkey/daemon.json
   lowkey start /path/to/watch
   ```

2. **Restore from backup** (if available):
   ```bash
   cp ~/backup/daemon.json ~/.local/state/lowkey/
   ```

### Stale state files

**Symptom**: Daemon behaves incorrectly after crash

**Solutions**:

```bash
# Clear all state and restart
lowkey clear --state --yes
lowkey start /path/to/watch
```

## Platform-Specific Issues

### macOS: "Operation not permitted"

**Cause**: Full Disk Access not granted

**Solution**:
1. Open System Preferences → Security & Privacy → Privacy
2. Click "Full Disk Access"
3. Click the lock to make changes
4. Add Terminal.app or your shell application

### Linux: systemd service fails

**Diagnosis**:

```bash
systemctl status lowkey.service
journalctl -u lowkey.service -n 50
```

**Common fixes**:
- Check paths in service file are absolute
- Verify binary exists and is executable
- Check permissions on state directory

### Windows: Service won't start

**Diagnosis**:

```powershell
# For NSSM service
nssm status lowkey

# Check Windows Event Viewer
Get-EventLog -LogName Application -Source lowkey -Newest 20
```

**Common fixes**:
- Verify paths use Windows format (`C:\path\to\...`)
- Check antivirus isn't blocking `lowkey.exe`
- Ensure NSSM service parameters are correct

## Getting Help

If issues persist after trying these solutions:

1. **Gather diagnostic information**:
   ```bash
   # Save output to file
   lowkey status > ~/lowkey-debug.txt
   lowkey tail >> ~/lowkey-debug.txt
   cat ~/.local/state/lowkey/daemon.json >> ~/lowkey-debug.txt
   ```

2. **Check GitHub issues**: [github.com/onedusk/lowkey/issues](https://github.com/onedusk/lowkey/issues)

3. **Open new issue** with:
   - Operating system and version
   - `lowkey --version` output
   - Steps to reproduce
   - Relevant log excerpts
   - Output from diagnostic commands

## Related Documentation

- [State Directories](state-directories.md) - Understanding state file locations
- [Telemetry Guide](telemetry.md) - Using metrics for diagnosis
- [Platform Setup Guides](macos-setup.md) - OS-specific daemon configuration
