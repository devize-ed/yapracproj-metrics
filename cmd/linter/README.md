# cmd/linter

This directory contains a linter that uses the static analyzer from the `internal/analyze` package.

## Overview

The linter is a command-line tool that checks Go code for forbidden patterns:
**Panic Detection**: Reports any usage of `panic` throughout the codebase
**Exit Detection**: Detects `os.Exit` and `log.Fatal` calls outside of `main()`

## Usage

```bash
# Build and run
go run ./cmd/linter ./...
```

### Example Output for forbiden usage:

```
/file.go:10:5: panic detected, usage of panic is forbidden
/file.go:25:8: log.Fatal detected outside of main function ; forbidden usage outside of main package
```

