#!/bin/bash

# Set the github environment to safe to avoid vcs error from GO
git config --global --add safe.directory /app

# Generate docs
make swag

# Run with air for hot-reload
exec air -c .air.toml
