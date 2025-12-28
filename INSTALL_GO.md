# Installing Go on Windows

## Quick Install

1. **Download Go**
   - Visit: https://go.dev/dl/
   - Download: `go1.21.x.windows-amd64.msi` (or latest 1.21+)

2. **Run Installer**
   - Double-click the MSI file
   - Follow the installation wizard
   - Default location is fine: `C:\Program Files\Go`

3. **Verify Installation**
   Open PowerShell or Command Prompt:
   ```powershell
   go version
   ```

   Should show something like:
   ```
   go version go1.21.x windows/amd64
   ```

4. **Set Up Workspace** (optional but recommended)
   ```powershell
   # Go will use these by default, but you can verify:
   go env GOPATH
   go env GOROOT
   ```

## Building BoltGuard

Once Go is installed:

```powershell
# In the boltguard directory
cd boltguard

# Download dependencies
go mod download

# Build it
go build -o bin/boltguard.exe ./cmd/boltguard

# Or use the Makefile (if you have make)
make build

# Run it
.\bin\boltguard.exe --version
```

## Quick Test

If you have Docker installed:

```powershell
# Pull a test image
docker pull alpine:latest

# Scan it
.\bin\boltguard.exe alpine:latest
```

## Troubleshooting

### "go: command not found" after install
- Close and reopen your terminal
- The installer adds Go to PATH, but existing terminals won't see it

### Build errors about missing modules
```powershell
go mod tidy
go mod download
```

### Can't find Docker daemon
- Make sure Docker Desktop is running
- Or test with a saved tarball:
  ```powershell
  docker save alpine:latest -o alpine.tar
  .\bin\boltguard.exe alpine.tar
  ```

## Next Steps

Once built, see [SETUP.md](SETUP.md) for usage examples.
