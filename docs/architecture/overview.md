# Lowkey Go Application Architecture

This document provides a high-level overview of the `lowkey` Go application, a command-line tool for file system monitoring. The application is built using the Cobra framework and is structured into three main packages: `cmd`, `internal`, and `pkg`.

## 1. `cmd` Package

This is the entry point of the application. It defines the CLI commands and their arguments.

- **`root.go`**: Sets up the main `lowkey` command and initializes configuration using Viper.
- **Command Files (`watch.go`, `start.go`, etc.)**: Each file defines a subcommand (e.g., `lowkey watch`). These files are responsible for parsing command-line arguments and calling the appropriate business logic in the `internal` package.

## 2. `internal` Package

This package contains the core business logic of the application.

- **`daemon`**: Manages the background monitoring process.
- **`watcher`**: Contains the logic for monitoring files and directories. This is the heart of the application.
- **`events`**: Provides platform-specific implementations for file system events.
- **`filters`**: Implements logic for ignoring files and directories.
- **`state`**: Manages the application's state, including cached file signatures and persisted manifests.
- **`logging`**: Handles log rotation and formatting.
- **`reporting`**: Aggregates and summarizes file change events.

## 3. `pkg` Package

This package contains shared libraries and utilities that are not specific to the core business logic.

- **`config`**: Handles loading and validation of configuration files.
- **`output`**: Provides functionality for rendering CLI output in different formats (e.g., plain text, JSON).
- **`telemetry`**: Contains hooks for metrics and tracing.

This architecture separates the command-line interface, core logic, and shared utilities, making the application modular and easier to maintain.

TODO: This document should be updated as the application evolves. The stubs in the `internal` and `pkg` directories need to be implemented to realize this architecture fully.
