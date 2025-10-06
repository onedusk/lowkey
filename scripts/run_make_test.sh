#!/bin/bash
# Wrapper script to run `make test` with a timeout.

set -euo pipefail

# 30-second timeout
TIMEOUT_SECONDS=30

# Run `make test` with the specified timeout.
# The `timeout` command is part of GNU coreutils.
# On macOS, it can be installed via `brew install coreutils`.
if command -v gtimeout &> /dev/null; then
    gtimeout $TIMEOUT_SECONDS make test
else
    timeout $TIMEOUT_SECONDS make test
fi
