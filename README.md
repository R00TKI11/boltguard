# BoltGuard ⚡

**BoltGuard** is a fast, offline-first container policy engine.

It evaluates container images against human-readable policies and answers one
question deterministically and early:

> **Does this image comply with our rules?**

BoltGuard is designed to run during local development, CI pipelines, and
air-gapped environments, with no required network access and no background
services.

---

## What BoltGuard is (and is not)

BoltGuard is a **policy engine for container images**.

It enforces rules such as:
- Must not run as root
- Must not use the `latest` tag
- Must include required metadata labels
- Must stay within size limits
- Must avoid forbidden files or paths

BoltGuard is intentionally **not** a full CVE scanner and does not maintain a
large vulnerability database. Instead, it focuses on deterministic, offline-
friendly policy enforcement.

BoltGuard is designed to **complement** tools like Trivy or Grype by acting as a
fast policy gate before heavier security analysis runs.

---

## Key features

- ⚡ Fast, deterministic scans
- 🔒 Offline-first by default
- 📜 Policy-as-code (YAML)
- 📦 Single static binary
- 🧱 Air-gap friendly (bundle import/export)
- 🧪 Designed for local development and CI
- 📤 Text, JSON, and SARIF output formats

---

## Quick start

Scan a local image using the default built-in policy:

    boltguard nginx:latest

Scan with a custom policy file:

    boltguard -policy policy.yaml alpine:3.18

Generate machine-readable output:

    boltguard -format json redis:7 > report.json

---

## Supported image inputs

BoltGuard can scan container images from multiple sources:

- Local images via a container runtime (for example: `nginx:latest`)
- Image tarballs created with `docker save`
- Offline policy bundles imported via `-bundle-import`

No network access is required during scanning unless explicitly enabled.

---

## Example policy

BoltGuard policies are intentionally simple and human-readable.

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

For details on policy structure and semantics, see
docs/policy.md.

---

## Output formats

BoltGuard supports multiple output formats:

- text (default, human-readable)
- json (automation and CI)
- sarif (GitHub code scanning)

Example:

    boltguard -format sarif nginx:latest > boltguard.sarif

---

## Air-gapped usage

BoltGuard is designed to operate in restricted and disconnected environments.

Policies can be bundled on a connected system and imported into an air-gapped
environment without network access.

On a connected system:

    boltguard -bundle-export policies.tar.gz

On the air-gapped system:

    boltguard -bundle-import policies.tar.gz
    boltguard myimage.tar

For more details, see docs/airgap.md.

---

## Performance

BoltGuard is designed to be fast by construction.

It avoids network calls, large vulnerability databases, and unnecessary
filesystem extraction. Scans operate on image metadata and selected file
signals only, allowing BoltGuard to run early in development workflows and CI.

Exact performance characteristics depend on image size, enabled policies,
and cache state. Formal benchmarks will be documented in a future release.

---

## Design philosophy

BoltGuard is built around a small set of non-negotiable constraints:

- Offline-first operation by default
- Deterministic and reproducible results
- Single static binary distribution
- Human-readable and auditable policies

For design details and non-goals, see docs/design.md.

---

## Installation

Prebuilt binaries are available on the GitHub releases page.

Alternatively, build from source:

    go build ./cmd/boltguard

---

## Contributing

Contributions are welcome.

Please see CONTRIBUTING.md for setup instructions,
development workflow, and contribution guidelines.

---

## License

BoltGuard is licensed under the MIT License.
