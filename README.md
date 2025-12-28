# BoltGuard ⚡

BoltGuard is a lightweight, policy-first container security tool.

## What BoltGuard is (and is not)

BoltGuard is a **policy engine for container images**.

It answers one question — quickly and deterministically:

> **Does this container image comply with our rules?**

BoltGuard is intentionally **not** a full CVE scanner and does not maintain a
large vulnerability database. Instead, it focuses on enforcing
human-readable, offline-friendly policies early in development and CI.

BoltGuard is designed to complement tools like Trivy or Grype by acting as a
fast, deterministic gate before heavier security analysis runs.

## Supported image inputs

BoltGuard can scan container images from multiple sources:

- Local images via a container runtime (e.g. `nginx:latest`)
- Image tarballs created with `docker save`
- Offline policy bundles imported via `-bundle-import`

No network access is required during scanning unless explicitly enabled.

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
rules:
  - id: require-non-root
    deny:
      user: root
    message: "Image must not run as root. Set USER in the Dockerfile."

  - id: disallow-latest-tag
    deny:
      tag: latest
    message: "Avoid using the 'latest' tag for reproducible builds."

  - id: require-standard-labels
    require:
      labels:
        - org.opencontainers.image.source
        - org.opencontainers.image.revision
    message: "Image must include standard OCI source and revision labels."

  - id: image-size-limit
    deny:
      image_size_mb_gt: 300
    message: "Image exceeds the maximum allowed size (300MB)."
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

## Performance

BoltGuard is designed to be fast by construction.

It avoids network calls, large vulnerability databases, and unnecessary
filesystem extraction. Scans operate on image metadata and selected file
signals only, allowing BoltGuard to run early in development workflows and CI.

Exact performance characteristics depend on image size, enabled policies,
and cache state. Formal benchmarks will be documented in a future release.

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
