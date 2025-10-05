# 1. Use Go for Development

*   **Status:** Accepted
*   **Date:** 2025-10-03

## Context and Problem Statement

We have a working prototype of `lowkey` in [Ruby](../.arc/lokee.rb) and [shell script](../.arc/lokee.sh). While these were excellent for rapid prototyping, they have limitations in terms of performance, cross-platform support, and maintainability. We need to choose a language for the production version of `lowkey` that is fast, reliable, and easy to maintain.

## Decision Drivers

*   **Performance:** The tool needs to be fast and efficient, especially when monitoring large directories with many files.
*   **Cross-Platform Support:** The tool must be easy to build and distribute for macOS, Windows, and Linux.
*   **Concurrency:** The tool needs to handle file system events concurrently without blocking the main thread.
*   **Maintainability:** The code should be easy to read, write, and refactor.
*   **Developer Experience:** The language should have a good ecosystem, community, and tooling.

## Considered Options

*   Go
*   Rust
*   Node.js
*   Python

## Decision Outcome

Chosen option: "Go", because it offers the best balance of performance, simplicity, and maintainability for this type of application. Go's built-in concurrency features (goroutines and channels) are a perfect fit for handling file system events, and its ability to compile to a single binary makes cross-platform distribution trivial.

### Positive Consequences

*   Excellent performance.
*   Easy to build and distribute for multiple platforms.
*   Clean and simple syntax.
*   Strong standard library.
*   Large and active community.

### Negative Consequences

*   Less memory safety than Rust (though still memory-safe).
*   Error handling can be verbose.
