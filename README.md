# BoltGuard ⚡

BoltGuard is a lightweight, policy-first container security tool.

It answers one question fast:

> **Is this container acceptable for *our* risk tolerance?**

BoltGuard runs in milliseconds, works fully offline, and uses
human-readable policies you can trust in development, CI, and air-gapped environments.

## Features

- **Fast** - Scans complete in milliseconds, not minutes
- **Offline-first** - No registry access or external databases required
- **Policy-driven** - Human-readable YAML policies you can customize
- **Multiple outputs** - Text for humans, JSON/SARIF for CI/CD
- **Air-gap friendly** - Works in restricted environments
- **Simple** - One binary, no dependencies

## Quick Start

```bash
# scan a local image
boltguard nginx:latest

# use custom policy
boltguard -policy policies/strict.yaml myapp:v1.2.3

# output as JSON for CI
boltguard -format json alpine:latest > report.json
```

## Installation

### From source

```bash
git clone https://github.com/R00TKI11/boltguard.git
cd boltguard
make build
./bin/boltguard --version
```

### Pre-built binaries

Download from [releases](https://github.com/R00TKI11/boltguard/releases)

## Usage

```
boltguard [options] <image>

Options:
  -policy string    Path to policy file (default: built-in)
  -format string    Output format: text, json, sarif (default "text")
  -v                Verbose output
  -version          Show version
  -offline          Offline mode (default true)
```

### Examples

```bash
# basic scan
boltguard nginx:latest

# strict production policy
boltguard -policy policies/strict.yaml prod-app:v2.1.0

# permissive dev policy
boltguard -policy policies/permissive.yaml dev-app:latest

# JSON output for automation
boltguard -format json myapp:v1 | jq '.summary'

# SARIF for GitHub Code Scanning
boltguard -format sarif myapp:v1 > results.sarif

# verbose mode
boltguard -v alpine:3.18
```

## Policies

BoltGuard ships with three example policies:

- **default.yaml** - Balanced for general use
- **strict.yaml** - High-security for production
- **permissive.yaml** - Lenient for development

### Example Policy

```yaml
name: "My Policy"
description: "Container security baseline"
version: "1.0.0"

settings:
  fail_on_error: false
  min_severity: "low"

rules:
  - id: "no-root"
    name: "Don't run as root"
    severity: "high"
    kind: "user"
    config:
      allow_root: false

  - id: "size-check"
    name: "Keep images lean"
    severity: "medium"
    kind: "size"
    config:
      max_mb: 2048

  - id: "no-secrets"
    name: "No hardcoded secrets"
    severity: "critical"
    kind: "env"
    config:
      deny_patterns:
        - "(?i)password"
        - "(?i)api[_-]?key"
```

See [docs/policy.md](docs/policy.md) for the full policy guide.

## What It Checks

BoltGuard evaluates images against your policy rules:

- **User/Root** - Is the container running as root?
- **Size** - Is the image unreasonably large?
- **Labels** - Are required metadata labels present?
- **Environment** - Are there hardcoded secrets in ENV vars?
- **Base Image** - Is it built from a trusted base?
- **Layers** - Does it have too many layers?

More rule types coming in future releases.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Container Policy Check
  run: |
    boltguard -format sarif myapp:${{ github.sha }} > results.sarif

- name: Upload results
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: results.sarif
```

### GitLab CI

```yaml
container-check:
  stage: verify
  script:
    - boltguard -format json $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA > report.json
  artifacts:
    reports:
      container_scanning: report.json
```

## Air-Gapped Usage

BoltGuard is designed for offline environments:

1. Build or download binary on connected machine
2. Transfer to air-gapped host
3. Run against local images (daemon or tarball)

No network access needed. See [docs/airgap.md](docs/airgap.md) for details.

## Documentation

- [Policy Guide](docs/policy.md) - Writing custom policies
- [Design Overview](docs/design.md) - Architecture and philosophy
- [Air-Gap Usage](docs/airgap.md) - Offline environments
- [Contributing](CONTRIBUTING.md) - How to contribute
- [Setup Guide](SETUP.md) - Installation and troubleshooting

## Why BoltGuard?

**Not another CVE scanner.** Tools like Trivy and Grype do vulnerability scanning well.

BoltGuard does something different: **fast, offline policy checks** you can run everywhere - dev laptops, CI, air-gapped production.

Think of it as a linter for your containers.

### Use BoltGuard for:
- Policy enforcement (no root, size limits, etc.)
- Quick sanity checks before deployment
- Air-gapped environments where CVE scanning isn't practical
- Development workflow (fast feedback loop)

### Use CVE scanners for:
- Finding known vulnerabilities
- Compliance reporting (SBOM, attestations)
- Deep package analysis

They complement each other.

## Roadmap

v0.2 ideas:
- [ ] Offline policy packs (curated policies for common stacks)
- [ ] Layer file inspection (check for setuid binaries, etc.)
- [ ] Custom evaluator plugins
- [ ] Policy composition/inheritance
- [ ] More output formats (JUnit, etc.)

See issues for details.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT - see [LICENSE](LICENSE)
