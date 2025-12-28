# BoltGuard Features

## Core Features (v0.1)

### Fast, Offline Scanning
- Scans complete in milliseconds
- No network access required
- Works in air-gapped environments
- Reads from local Docker daemon or tarballs

### Policy-Driven Checks
- Human-readable YAML policies
- Extensible rule system
- Built-in rule types:
  - User/root checks
  - Image size limits
  - Required labels
  - Environment variable scanning (secrets detection)
  - Base image validation
  - Layer count optimization

### Multiple Output Formats
- **Text** - Human-readable console output
- **JSON** - Machine-readable for CI/CD
- **SARIF** - GitHub Code Scanning integration

### Caching for Speed
```bash
# Automatic caching by image digest
boltguard alpine:latest          # first run: full scan
boltguard alpine:latest          # second run: uses cache

# Cache management
boltguard -cache-stats           # view cache statistics
boltguard -cache-clear           # clear all cached results
boltguard -cache-dir /tmp/cache  # custom cache location
boltguard -cache=false nginx     # disable cache for one scan
```

### Bundle System for Air-Gap
```bash
# Export policies as portable bundle
boltguard -bundle-export policies.tar.gz

# Transfer to air-gapped host and import
boltguard -bundle-import policies.tar.gz

# List installed bundles
boltguard -bundle-list
```

Perfect for:
- Regulated industries
- Defense/government
- Offline development
- Compliance requirements

## Built-in Policies

### default.yaml
Balanced policy for general use:
- No root user (high severity)
- 2GB size limit (medium)
- Basic label requirements (low)
- Secret detection in ENV (critical)
- Known base image check (medium)
- 50 layer limit (low)

### strict.yaml
High-security for production:
- No root (critical, fail-fast)
- 1GB size limit (high)
- Extensive label requirements (high)
- Strict secret patterns (critical)
- Only approved bases (critical)
- 20 layer limit (medium)

### permissive.yaml
Lenient for development:
- Root allowed but warned (low)
- 3GB size limit (low)
- Minimal checks
- Focus on obvious issues only

## CLI Usage

### Basic Scanning
```bash
# Scan with default policy
boltguard nginx:latest

# Custom policy
boltguard -policy my-policy.yaml app:v1.2.3

# Verbose output
boltguard -v alpine:latest

# Different output formats
boltguard -format json redis:7 > report.json
boltguard -format sarif app:v1 > results.sarif
```

### Cache Operations
```bash
# View cache stats
boltguard -cache-stats

# Clear cache
boltguard -cache-clear

# Disable cache for single scan
boltguard -cache=false nginx:latest

# Custom cache directory
boltguard -cache-dir /tmp/mycache alpine
```

### Bundle Management
```bash
# Import policy bundle
boltguard -bundle-import security-policies.tar.gz

# List installed bundles
boltguard -bundle-list

# Export current policies
boltguard -bundle-export my-policies.tar.gz
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Container Security Check
  run: |
    boltguard -format sarif ${{ env.IMAGE }} > results.sarif

- name: Upload to Code Scanning
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: results.sarif
```

### GitLab CI
```yaml
container-policy:
  stage: security
  script:
    - boltguard -format json $IMAGE > report.json
  artifacts:
    reports:
      container_scanning: report.json
```

### Jenkins
```groovy
stage('Container Policy') {
    steps {
        sh 'boltguard -format json myapp:${BUILD_NUMBER} > report.json'
        archiveArtifacts 'report.json'
    }
}
```

## Advanced Usage

### Custom Policies
Create `my-policy.yaml`:
```yaml
name: "My Custom Policy"
version: "1.0.0"

settings:
  fail_on_error: true
  min_severity: "medium"

rules:
  - id: "custom-check"
    name: "My custom rule"
    severity: "high"
    kind: "user"
    config:
      allow_root: false
```

### Policy from Bundle
```bash
# After importing a bundle
boltguard -policy ~/.config/boltguard/packs/security-bundle/strict.yaml nginx
```

### Scanning Tarballs
```bash
# Save image as tarball
docker save myapp:v1 -o myapp.tar

# Scan the tarball (useful for air-gap)
boltguard myapp.tar
```

## Performance

Typical scan times:
- Small image (alpine): ~50-100ms
- Medium image (nginx): ~100-200ms
- Large image (node): ~200-500ms

With caching:
- Cached scan: ~10-50ms (up to 10x faster)

## What's NOT Included

BoltGuard focuses on policy checks, not:
- ❌ CVE scanning (use Trivy/Grype)
- ❌ SBOM generation (use Syft)
- ❌ Runtime protection (use Falco)
- ❌ Image signing (use Sigstore)

These tools complement BoltGuard perfectly.

## Roadmap (v0.2+)

Planned features:
- [ ] Full cache reuse (skip re-scanning unchanged images)
- [ ] Layer file extraction (check for setuid binaries, etc.)
- [ ] Custom rule plugins (Go plugins or WASM)
- [ ] Policy inheritance/composition
- [ ] More output formats (JUnit, HTML)
- [ ] Curated policy packs (for Node.js, Python, etc.)
- [ ] Advisory system (base image recommendations)
- [ ] Policy signing and verification
- [ ] Multi-image batch scanning

## Library Usage

BoltGuard can be embedded in other Go tools:

```go
import (
    "github.com/R00TKI11/boltguard/internal/image"
    "github.com/R00TKI11/boltguard/internal/facts"
    "github.com/R00TKI11/boltguard/internal/policy"
    "github.com/R00TKI11/boltguard/internal/rules"
)

// Load image
img, _ := image.Load("nginx:latest", true)

// Extract facts
f, _ := facts.Extract(img)

// Load policy
pol, _ := policy.LoadDefault()

// Evaluate
engine := rules.NewEngine()
results := engine.Evaluate(f, pol)

// Use results...
```

## Configuration

BoltGuard looks for policies in:
1. `-policy` flag
2. `policies/default.yaml` (project)
3. `/etc/boltguard/default.yaml` (system)
4. `~/.config/boltguard/default.yaml` (user)
5. Embedded default

Cache location:
- Default: `~/.cache/boltguard`
- Override: `-cache-dir`

Bundle storage:
- Default: `~/.config/boltguard/packs`

## Exit Codes

- `0` - All checks passed
- `1` - One or more checks failed OR error occurred

Check JSON/SARIF output for detailed results.
