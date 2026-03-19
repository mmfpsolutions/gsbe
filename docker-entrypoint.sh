#!/bin/sh
set -e

# Create required directories
mkdir -p /app/config /app/logs

# Ensure correct ownership
chown -R app:app /app/config /app/logs

# Run as app user
exec su-exec app /app/gsbe "$@"
