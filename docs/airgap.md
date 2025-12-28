# Air-Gapped Usage

BoltGuard is designed to work fully offline, making it ideal for air-gapped or restricted environments.

## How It Works

BoltGuard only analyzes what's already on the host:

1. **No registry access needed** - reads from local Docker daemon or tarball
2. **No external databases** - all checks are policy-based
3. **No phone-home** - zero telemetry or analytics

### Typical air-gapped workflow

On a machine with internet access:

    boltguard -bundle-export policies.tar.gz

Transfer the bundle using removable media or an internal artifact store.

On the air-gapped system:

    boltguard -bundle-import policies.tar.gz
    boltguard myimage.tar

## Air-Gapped Setup

### Option 1: Build from source

On an internet-connected machine:

```bash
# download dependencies
go mod download
go mod vendor

# build binary
go build -o boltguard ./cmd/boltguard

# copy binary + policies to air-gapped host
```

### Option 2: Pre-built binary

Download the release binary and transfer to air-gapped environment.

### Option 3: Import via tarball

Save images as tarballs on connected machine:

```bash
docker pull nginx:latest
docker save -o nginx-latest.tar nginx:latest
```

Transfer tarball to air-gapped host:

```bash
docker load -i nginx-latest.tar
boltguard nginx:latest
```

### What is included in a bundle

A BoltGuard bundle contains:

- Policy definitions
- Metadata required for evaluation
- Versioned schema information

Bundles are intentionally explicit and self-contained. BoltGuard does not
automatically fetch updates or external data when operating offline.

## CI/CD in Air-Gap

BoltGuard fits naturally into offline CI:

```yaml
# GitLab CI example
image-check:
  stage: verify
  script:
    - boltguard -policy /etc/policies/prod.yaml $IMAGE_NAME
    - boltguard -format json $IMAGE_NAME > report.json
  artifacts:
    reports:
      container_scanning: report.json  # or use sarif
```

## Offline Policy Distribution

To distribute policies in air-gap environments:

1. **Version control** - commit policies to git
2. **Config management** - deploy via Ansible/Chef/Puppet
3. **Artifact registry** - store in internal registry
4. **Embedded** - use the built-in default policy

## Limitations

Without network access, BoltGuard can't:

- Pull images from remote registries (use local daemon or tarballs)
- Fetch CVE databases (not its job anyway)
- Check if base images are outdated (policy can enforce specific versions)

## Complementary Tools

In air-gap, combine BoltGuard with:

- **Trivy** (offline mode) - CVE scanning with local DB
- **Clair** - Vulnerability analysis with synced DB
- **Anchore** - Deeper analysis with offline feeds

BoltGuard handles fast policy checks, let other tools handle CVEs.

## Example: Full Offline Workflow

```bash
# 1. Export policy from secure environment
git clone internal-registry/boltguard-policies
cd boltguard-policies

# 2. Scan locally built image
boltguard -policy prod-strict.yaml myapp:v2.1.0

# 3. Generate compliance report
boltguard -format sarif -policy prod-strict.yaml myapp:v2.1.0 > report.sarif

# 4. Archive for audit
tar czf scan-$(date +%Y%m%d).tar.gz report.sarif
```

## Security Considerations

Air-gap doesn't mean "trust everything local":

- **Still validate images** - they could be misconfigured
- **Still check for secrets** - leaked credentials are leaked regardless
- **Still enforce least privilege** - running as root is always risky

BoltGuard helps enforce policy even when you can't check CVEs online.
