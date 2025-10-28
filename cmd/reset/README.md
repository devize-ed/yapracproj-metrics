# cmd/reset

This directory contains a code generation tool that automatically generates `Reset()` methods for structs.

The reset tool scans Go source files for structs annotated with the `// generate:reset` comment and generates corresponding `Reset()` methods that reset all fields to their zero values.

## Usage

Use the `// generate:reset` comment above the struct for generating

```bash
# From the project root
go run ./cmd/reset

```

The tool generates a file named `{StructName}_reset.go` next to the source file



