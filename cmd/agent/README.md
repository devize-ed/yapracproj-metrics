# cmd/agent

This directory contains the agent application entry point.

## Overview

The agent application provides system metrics collection and transmission to server


### Environment variables

- `ADDRESS`: Server address (default: localhost:8080)
- `POLL_INTERVAL`: Metric collection interval (seconds)
- `REPORT_INTERVAL`: Metric transmission interval (seconds)
- `RATE_LIMIT`: Number of concurrent workers
- `ENABLE_GZIP`: Enable compression for requests
- `ENABLE_GET_METRICS`: Enable test mode for metric retrieval
- `KEY`: Secret key for request signing

## Command-line flags

```bash
# Start with custom configuration
./agent -a localhost:8080 -r 10 -p 2 -l 4 -c -k "secret-key"
```
