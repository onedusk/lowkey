# Project Overview

This directory contains "Lokee", a file monitoring tool designed to track changes (creations, modifications, deletions) in specified directories. The tool is implemented in two versions: a Ruby script (`lokee.rb`) and a shell script (`lokee.sh`). Both versions provide functionalities to watch directories in the foreground, run monitoring as a background process, and view logs of file changes.

The core of the project is a file monitoring algorithm that periodically scans directories and compares file states (mtime and size) to detect changes. The accompanying markdown documents, `doc.md` and `lokee_algorithm_analysis.md`, provide a deep dive into the algorithm's complexity and propose several optimization strategies, including:

* A hybrid approach using filesystem events and polling.
* Incremental state tracking to reduce redundant checks.
* Smart filtering with Bloom filters.
* Adaptive polling with exponential backoff.
* Batch processing of changes.
* Memory-efficient state storage.
* A priority queue for processing changes.

# Building and Running

The scripts are self-contained and do not require a separate build process.

## Ruby Version (`lokee.rb`)

The Ruby version is a more feature-rich implementation using the `thor` gem for command-line interface management.

**Key Commands:**

* **Watch directories in the foreground:**

    ```bash
    ./lokee.rb watch <dir1> <dir2> ...
    ```

* **Start background monitoring:**

    ```bash
    ./lokee.rb start <dir1> <dir2> ...
    ```

* **Stop background monitoring:**

    ```bash
    ./lokee.rb stop
    ```

* **View monitor status:**

    ```bash
    ./lokee.rb status
    ```

* **View logs:**

    ```bash
    ./lokee.rb log [pattern]
    ```

* **Follow logs in real-time:**

    ```bash
    ./lokee.rb tail
    ```

* **Show change statistics:**

    ```bash
    ./lokee.rb summary
    ```

* **Clear all logs:**

    ```bash
    ./lokee.rb clear
    ```

## Shell Version (`lokee.sh`)

The shell version provides similar functionalities to the Ruby version but is implemented as a bash script.

**Key Commands:**

* **Watch directories in the foreground:**

    ```bash
    ./lokee.sh watch <dir1> <dir2> ...
    ```

* **Start background monitoring:**

    ```bash
    ./lokee.sh start <dir1> <dir2> ...
    ```

* **Stop background monitoring:**

    ```bash
    ./lokee.sh stop
    ```

* **View monitor status:**

    ```bash
    ./lokee.sh status
    ```

* **View logs:**

    ```bash
    ./lokee.sh log [pattern]
    ```

* **Follow logs in real-time:**

    ```bash
    ./lokee.sh tail
    ```

* **Show change statistics:**

    ```bash
    ./lokee.sh summary
    ```

* **Clear all logs:**

    ```bash
    ./lokee.sh clear
    ```

# Development Conventions

* **Ruby:** The Ruby script follows standard Ruby conventions and is well-documented with inline comments. It uses the `thor` gem for building the CLI.
* **Shell:** The shell script is written in bash and is also well-documented. It uses standard shell commands and features.
* **Algorithm Analysis:** The markdown files provide a detailed mathematical analysis of the algorithm and its potential optimizations. This suggests a focus on performance and efficiency.
