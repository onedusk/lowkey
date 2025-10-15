# Product Requirements Document: Lowkey

## 1. Introduction

Lowkey is a high-performance, cross-platform file monitoring tool for developers and power users. It provides a simple and efficient way to watch directories for changes and log them.

## 2. Vision and Goals

### 2.1. Vision

To provide a simple, reliable, and high-performance file monitoring tool that is simple to use. **TIP**: can be used to automate documentation and table of contents creation for [jot](https://github.com/onedusk/jot)

### 2.2. Goals

*   **Performance:** Lowkey should be fast and have a low memory footprint.
*   **Reliability:** Lowkey should be stable and reliable, even when monitoring large and active directories.
*   **Ease of Use:** Lowkey should be easy to install, configure, and use.
*   **Cross-Platform:** Lowkey should work on macOS, Windows, and Linux.

## 3. Features

### 3.1. Core Features

*   Monitor one or more directories for file changes (create, modify, delete).
*   Log changes to a file and rotate when logs exceed 10 MB (configurable).
*   Ignore files and directories using a `.lowkey` file.
*   Run interactively in the foreground or supervise a single background agent per host.

### 3.2. CLI Commands

*   `lowkey watch <dirs...>`: Watch directories in the foreground until interrupted.
*   `lowkey start <dirs...>`: Launch a background daemon (one per user) and persist watch targets to `$XDG_STATE_HOME/lowkey/daemon.json`.
*   `lowkey stop`: Stop the background daemon via PID file and clear transient state.
*   `lowkey status`: Report daemon PID, configured directories, and last heartbeat timestamp.
*   `lowkey log [--since duration]`: Print rotated log files, filtered by pattern or time window.
*   `lowkey tail`: Stream the live daemon log.
*   `lowkey summary`: Show change statistics aggregated from the state store.
*   `lowkey clear [--logs|--state]`: Prune logs, state, or both with confirmation.

## 4. User Stories

*   As a developer, I want to monitor my source code for changes so that I can automatically trigger a build or run tests.
*   As a system administrator, I want to monitor configuration files for changes so that I can be alerted to unauthorized modifications.
*   As a power user, I want to monitor a directory for new files so that I can automatically process them.
*   As an SRE, I want the daemon to restart cleanly on reboot so that monitoring resumes without manual intervention.

## 5. Design and UX

Lowkey will be a command-line tool with a simple and intuitive interface. Foreground commands print terse status lines, while daemon-oriented commands emit structured tables and machine-readable JSON when `--output json` is passed. Color-coded logs aid readability, and lifecycle commands surface actionable errors (e.g., missing PID file, permission denied). Daemon state lives in `$XDG_STATE_HOME/lowkey/` (macOS fallback `~/Library/Application Support/lowkey`, Windows `%LOCALAPPDATA%\lowkey`).

## 6. Technical Requirements

*   **Language:** Go
*   **Dependencies:** `fsnotify` for file system notifications.
*   **Process Supervision:** Daemon writes PID + heartbeat file to platform-specific state directory and uses OS-native signals (`SIGTERM`, `CTRL_BREAK_EVENT`, `CancelIoEx`) for shutdown.
*   **State & Config:** Watch manifests stored as JSON in `$XDG_STATE_HOME/lowkey/daemon.json` with atomic updates.
*   **Logging:** Structured JSON logs rotated at 10 MB, retained up to five files.
*   **Platform:** macOS, Windows, Linux

## 7. Future Work

*   JSON output for logs.
*   Webhooks to notify other services of changes.
*   Plugin system for extending functionality.
*   Optional systemd/service wrappers for unattended installs.
