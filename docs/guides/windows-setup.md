# Windows Setup Guide

This guide covers installation and background service setup for lowkey on Windows.

## Installation

Build and install the lowkey binary:

```powershell
# Build the binary (requires Go installed)
go build -o lowkey.exe ./cmd/lowkey

# Move to a directory in PATH
# Option 1: System-wide (requires Administrator)
Move-Item lowkey.exe C:\Windows\System32\

# Option 2: User-specific
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\Programs\lowkey"
Move-Item lowkey.exe "$env:LOCALAPPDATA\Programs\lowkey\"
# Add to PATH: System Properties > Environment Variables > User Variables > Path
```

## Configure as Windows Service

### Option 1: Using NSSM (Non-Sucking Service Manager)

Download and install NSSM from [nssm.cc](https://nssm.cc/download):

```powershell
# Install lowkey as a service (run as Administrator)
nssm install lowkey "C:\Windows\System32\lowkey.exe" start --metrics 127.0.0.1:9600 --trace

# Configure service properties
nssm set lowkey AppDirectory "C:\ProgramData\lowkey"
nssm set lowkey AppStdout "C:\ProgramData\lowkey\logs\lowkey-stdout.log"
nssm set lowkey AppStderr "C:\ProgramData\lowkey\logs\lowkey-stderr.log"
nssm set lowkey DisplayName "Lowkey Filesystem Monitor"
nssm set lowkey Description "High-performance filesystem monitoring daemon"
nssm set lowkey Start SERVICE_AUTO_START

# Start the service
nssm start lowkey
```

### Option 2: Using Task Scheduler

For users without Administrator access:

```powershell
# Create scheduled task XML
$action = New-ScheduledTaskAction -Execute "lowkey.exe" -Argument "start --metrics 127.0.0.1:9600 --trace"
$trigger = New-ScheduledTaskTrigger -AtLogOn
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
$principal = New-ScheduledTaskPrincipal -UserId "$env:USERNAME" -LogonType Interactive -RunLevel Limited

# Register the task
Register-ScheduledTask -TaskName "Lowkey Monitor" -Action $action -Trigger $trigger -Settings $settings -Principal $principal -Description "Lowkey filesystem monitoring daemon"
```

## Managing the Service

### NSSM Service Management

```powershell
# Start service
nssm start lowkey

# Stop service
nssm stop lowkey

# Restart service
nssm restart lowkey

# Check status
nssm status lowkey

# Remove service
nssm remove lowkey confirm
```

### Task Scheduler Management

```powershell
# Start task
Start-ScheduledTask -TaskName "Lowkey Monitor"

# Stop task
Stop-ScheduledTask -TaskName "Lowkey Monitor"

# Get task status
Get-ScheduledTask -TaskName "Lowkey Monitor" | Get-ScheduledTaskInfo

# Remove task
Unregister-ScheduledTask -TaskName "Lowkey Monitor" -Confirm:$false
```

## Verify Status

```powershell
# Check daemon status
lowkey status

# View logs
lowkey tail
```

## Log Files

Lowkey creates log files in these locations:

### NSSM Service Logs
- **Stdout**: `C:\ProgramData\lowkey\logs\lowkey-stdout.log`
- **Stderr**: `C:\ProgramData\lowkey\logs\lowkey-stderr.log`
- **Daemon log**: View with `lowkey tail`

### Task Scheduler Logs
- **Application logs**: Check Windows Event Viewer under Task Scheduler
- **Daemon log**: Default location varies by user, check `lowkey status` for path

## Troubleshooting

### Service fails to start

Check the service status:
```powershell
# For NSSM
nssm status lowkey

# For Task Scheduler
Get-ScheduledTask -TaskName "Lowkey Monitor" | Get-ScheduledTaskInfo
```

View error logs:
```powershell
# NSSM logs
Get-Content C:\ProgramData\lowkey\logs\lowkey-stderr.log -Tail 50

# Task Scheduler logs
Get-WinEvent -LogName "Microsoft-Windows-TaskScheduler/Operational" -MaxEvents 50 | Where-Object { $_.Message -like "*Lowkey*" }
```

### Permission denied errors

Ensure the binary has execute permissions:
```powershell
# Check current permissions
icacls lowkey.exe

# Grant execute permissions
icacls lowkey.exe /grant Everyone:RX
```

Verify log directory exists:
```powershell
New-Item -ItemType Directory -Force -Path "C:\ProgramData\lowkey\logs"
```

### Service not starting at boot

For NSSM service:
```powershell
# Verify startup type
nssm set lowkey Start SERVICE_AUTO_START
```

For Task Scheduler:
```powershell
# Verify trigger settings
Get-ScheduledTask -TaskName "Lowkey Monitor" | Select-Object -ExpandProperty Triggers
```

### Firewall blocking metrics endpoint

If using `--metrics`, allow the port through Windows Firewall:
```powershell
# Allow inbound traffic on metrics port (requires Administrator)
New-NetFirewallRule -DisplayName "Lowkey Metrics" -Direction Inbound -Protocol TCP -LocalPort 9600 -Action Allow
```

## Notes

- **NSSM** is recommended for production deployments as it provides robust service management
- **Task Scheduler** is suitable for single-user scenarios without Administrator privileges
- Ensure antivirus software doesn't block lowkey.exe
- Check Windows Event Viewer for additional diagnostic information
