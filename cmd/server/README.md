# cmd/server

This directory contains the server application entry point.

## Overview

The server application provides HTTP server for metric storage and retrieval

### Environment variables

- `ADDRESS`: Server listen address (default: localhost:8080)
- `DATABASE_DSN`: Database connection string
- `FILE_PATH`: File storage path
- `STORE_INTERVAL`: File save interval (seconds)
- `KEY`: Secret key for request signing
- `AUDIT_FILE`: Audit log file path
- `AUDIT_URL`: Audit log URL endpoint

## Command-line flags

```bash
# Start with custom configuration
./server -a localhost:8080 -d "postgres://user:pass@localhost/db" -k "secret-key"
```
