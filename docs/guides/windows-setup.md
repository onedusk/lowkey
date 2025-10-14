# Windows Setup Guide

This guide explains how to install `lowkey` on Windows and configure it to run as a background service using the Non-Sucking Service Manager (NSSM).

## Installation

You can install `lowkey` by building from source or by downloading a pre-compiled binary.

### From Source

Ensure you have Go 1.22+ installed.

```powershell
# Clone the repository
git clone https://github.com/onedusk/lowkey.git
cd lowkey

# Build the binary
go build -o lowkey.exe ./cmd/lowkey

# Move the executable to a permanent location
New-Item -ItemType Directory -Force -Path "C:\Program Files\lowkey"
Move-Item -Path .\lowkey.exe -Destination "C:\Program Files\lowkey\"
```

### From Binary

Download the latest `lowkey.exe` from the project's releases page.

```powershell
# Create a directory for the application
New-Item -ItemType Directory -Force -Path "C:\Program Files\lowkey"

# Download the binary (example URL)
Invoke-WebRequest -Uri "https://github.com/onedusk/lowkey/releases/latest/download/lowkey-windows-amd64.exe" -OutFile "C:\Program Files\lowkey\lowkey.exe"
```

Finally, add `C:\Program Files\lowkey` to your system's `Path` environment variable.

## Daemon Setup with NSSM

NSSM is a robust tool for running applications as Windows services.

### 1. Download NSSM

Download the latest release of NSSM from [nssm.cc](https://nssm.cc/download) and extract it. Copy `nssm.exe` (for your system's architecture) to a directory in your `Path`.

### 2. Install the Service

Open a PowerShell terminal as an **Administrator**.

```powershell
# Install the lowkey service
nssm install lowkey "C:\Program Files\lowkey\lowkey.exe"

# Set the service arguments
nssm set lowkey AppParameters "start --metrics 127.0.0.1:9600 C:\path\to\monitor"

# Set service details
nssm set lowkey DisplayName "Lowkey Monitor"
nssm set lowkey Description "A service that monitors filesystem events."
nssm set lowkey Start SERVICE_AUTO_START

# (Optional) Configure log redirection
nssm set lowkey AppStdout "C:\ProgramData\lowkey\logs\service.log"
nssm set lowkey AppStderr "C:\ProgramData\lowkey\logs\error.log"
nssm set lowkey AppRotateFiles 1
nssm set lowkey AppRotateBytes 10485760 # 10 MB

# Start the service
nssm start lowkey
```

**Note:** Replace `C:\path\to\monitor` with the directory you want `lowkey` to watch.

## Verification

Verify that the `lowkey` service is running correctly.

### Check Service Status

```powershell
# Check the service status with NSSM
nssm status lowkey
# Expected output: SERVICE_RUNNING

# Or use PowerShell's Get-Service
Get-Service -Name lowkey
```

### Check `lowkey` Status

Use the `lowkey` CLI to check the daemon's internal status.

```powershell
lowkey status
```

### View Logs

If you configured log redirection with NSSM, you can view the log files directly.

```powershell
# View the last 20 lines of the service log
Get-Content "C:\ProgramData\lowkey\logs\service.log" -Tail 20

# Tail logs using the built-in lowkey command
lowkey tail
```
