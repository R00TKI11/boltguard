# BoltGuard Setup Guide

## Prerequisites

You'll need:

1. **Go 1.21+** - [Download from golang.org](https://go.dev/dl/)
2. **Docker** - For testing against real images
3. **Make** (optional) - Makes building easier

## Installing Go (Windows)

1. Download Go installer from https://go.dev/dl/
2. Run the installer (e.g., `go1.21.windows-amd64.msi`)
3. Verify installation:
   ```powershell
   go version
   ```

## Building BoltGuard

### With Make

```bash
# download dependencies
make deps

# build the binary
make build

# run tests
make test

# install to $GOPATH/bin
make install
```

### Without Make

```bash
# download dependencies
go mod download

# build
go build -o bin/boltguard.exe ./cmd/boltguard

# or on Linux/Mac
go build -o bin/boltguard ./cmd/boltguard

# run tests
go test ./...

# install
go install ./cmd/boltguard
```

## First Run

Test against a local image:

```bash
# pull a test image
docker pull alpine:latest

# scan it
./bin/boltguard alpine:latest

# or if installed
boltguard alpine:latest
```

You should see output like:

```
BoltGuard Report
================

Image:    alpine:latest
Policy:   BoltGuard Default Policy (v0.1.0)
Scanned:  2025-12-28T...

Summary
-------
Total checks: 6
  Passed:     4
  Failed:     2
...
```

## Using Custom Policies

```bash
# use a stricter policy
boltguard -policy policies/strict.yaml nginx:latest

# permissive for dev images
boltguard -policy policies/permissive.yaml myapp:dev

# output as JSON
boltguard -format json alpine:latest > report.json

# SARIF for CI integration
boltguard -format sarif myapp:prod > report.sarif
```

## Troubleshooting

### "go: command not found"

Go is not installed or not in PATH. Install from https://go.dev/dl/

### "Cannot connect to Docker daemon"

Docker isn't running. Start Docker Desktop or the Docker daemon.

### "Image not found"

The image doesn't exist locally. Either:
- Pull it: `docker pull <image>`
- Build it: `docker build -t <image> .`
- Use a tarball: `boltguard /path/to/image.tar`

### Build errors

Make sure you're on Go 1.21+:
```bash
go version
```

If you see dependency issues:
```bash
go mod tidy
go mod download
```

## Next Steps

- Read [docs/policy.md](docs/policy.md) to write custom policies
- Check [docs/design.md](docs/design.md) to understand the architecture
- See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to contribute

## Quick Reference

```bash
# help
boltguard --help

# version
boltguard --version

# verbose output
boltguard -v nginx:latest

# custom policy
boltguard -policy my-policy.yaml app:v1

# output formats
boltguard -format text  app:v1  # default, human-readable
boltguard -format json  app:v1  # machine-readable
boltguard -format sarif app:v1  # CI/CD integration
```
