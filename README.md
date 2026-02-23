# arch-lint

A Go-based architecture linter that enforces dependency rules between package groups. Define which packages can (or cannot) depend on each other via a simple YAML configuration, and catch architectural violations early.

## Overview

Large Go codebases tend to develop unintended dependency chains over time. `arch-lint` lets you define logical **groups** of packages and specify explicit **allow/deny** dependency rules between them. It parses Go source files, resolves imports, and checks them against your rules.

## Installation

```bash
go install github.com/coderhyme/arch-lint/cmd/arch-lint@latest
```

## Configuration

Create a `.arch-lint.yaml` in your project root (see [`.arch-lint.example.yaml`](.arch-lint.example.yaml) for a full example):

```yaml
version: 1

groups:
  shared:
    paths:
      - "shared/x"

  core:
    paths:
      - "internal/core"
    dependencies:
      allow:
        groups:
          - shared          # can import packages in the "shared" group
        patterns:
          - "pkg/utils/*"   # glob pattern matching
        relative:
          - ../helpers      # relative path from the package
        subPackages: true   # allow importing own sub-packages
      deny:
        groups:
          - api             # cannot import packages in the "api" group
        patterns:
          - "legacy/*"      # block legacy packages
```

### Configuration Reference

| Field | Description |
|---|---|
| `version` | Config version (currently `1`) |
| `groups` | Map of group names to their definitions |
| `groups.<name>.paths` | Package paths belonging to this group (string, object, or array) |
| `groups.<name>.dependencies.allow` | Rules for allowed imports |
| `groups.<name>.dependencies.deny` | Rules for denied imports |

### Dependency Rule Types

| Rule | Description |
|---|---|
| `groups` | Reference other defined groups by name |
| `patterns` | Glob patterns matched against import paths |
| `relative` | Relative paths resolved from the importing package |
| `subPackages` | When `true`, allows importing own sub-packages |

### Path Configuration

Paths support multiple formats:

```yaml
# Simple string
paths: "pkg/foo"

# Array of strings
paths:
  - "pkg/foo"
  - "pkg/bar"

# Detailed object with filtering
paths:
  - dir: "pkg/foo"
    exclude: ["pkg/foo/internal"]
    include: ["*.go"]
    recursive: true
```

### Glob Pattern Behavior

Paths use [gobwas/glob](https://github.com/gobwas/glob) for matching. Note that `**` only matches **sub-paths**, not the directory itself:

```yaml
# Matches only "internal/domain"
paths: "internal/domain"

# Matches sub-packages only (NOT "internal/domain" itself)
paths: "internal/domain/**"

# Matches both "internal/domain" and all sub-packages
paths: "internal/domain{,/**}"
```

## Usage

```bash
arch-lint [flags]
```

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `.arch-lint.yaml` | Path to the configuration file |

```bash
# Run with default .arch-lint.yaml in current directory
arch-lint

# Specify a config file
arch-lint --config path/to/.arch-lint.yaml
```

### CI Integration

Add `arch-lint` to your CI pipeline to prevent architectural drift:

```yaml
# GitHub Actions
- name: Architecture lint
  run: |
    go install github.com/coderhyme/arch-lint/cmd/arch-lint@latest
    arch-lint
```

The process exits with code `1` when violations are found, making it suitable for CI gates.

## How It Works

1. **Parse config** - Reads the YAML configuration and validates group definitions
2. **Build groups** - Creates path matchers (glob-based) for each group and resolves inter-group references
3. **Scan packages** - Walks the Go source tree, parses imports from each `.go` file (excluding tests and vendor)
4. **Check rules** - For each package, determines its group membership and validates all imports against allow/deny rules
   - Deny rules are evaluated first
   - Allow rules are evaluated second
   - An import must pass all deny rules (not be denied) and match at least one allow rule

## Project Structure

```
.
├── cmd/arch-lint/       # CLI entrypoint
├── internal/
│   ├── checker/         # Violation detection
│   │   └── checker.go
│   ├── config/          # YAML config parsing and validation
│   │   ├── reader.go    # File loading and validation
│   │   └── types.go     # Config type definitions
│   ├── groups/          # Group management and dependency checking
│   │   ├── builder.go   # Group construction from config
│   │   ├── group.go     # Group and DependencyChecker interfaces
│   │   ├── import_rule.go  # Import rule implementations
│   │   ├── manager.go   # GroupManager implementation
│   │   └── path_matcher.go # Glob-based path matching
│   └── loader/          # Go source file traversal and import extraction
│       └── parser.go    # Go import parser
├── .arch-lint.example.yaml  # Example configuration
└── go.mod
```

## License

MIT
