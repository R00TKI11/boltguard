# Policy Guide

## Policy Structure

A BoltGuard policy is a YAML file with this structure:

```yaml
name: "My Policy"
description: "What this policy enforces"
version: "1.0.0"

settings:
  fail_on_error: false      # fail whole check if a rule errors?
  min_severity: "low"       # ignore issues below this level

rules:
  - id: "unique-rule-id"
    name: "Human readable name"
    description: "What this rule checks"
    severity: "high"        # critical, high, medium, low, info
    kind: "user"            # rule type
    fail_fast: false        # stop immediately if this fails?
    config:                 # rule-specific settings
      key: value
```

## Policy stability guarantees

BoltGuard policies are versioned and evaluated deterministically.

Changes to the policy schema are introduced explicitly and will not silently
change the meaning of existing policies. Older policy files will continue to
work or fail loudly if incompatible.

### Why BoltGuard policies are intentionally simple

BoltGuard policies are designed to be readable, auditable, and predictable.

Unlike general-purpose policy languages, BoltGuard avoids embedding complex
logic or control flow. This keeps policies easy to review and suitable for
restricted or regulated environments.

## Rule Types

### `user` - User/Root Checks

Validates which user the container runs as.

**Config:**
- `allow_root` (bool) - whether root is acceptable

**Example:**
```yaml
- id: "no-root"
  name: "Don't run as root"
  severity: "high"
  kind: "user"
  config:
    allow_root: false
```

### `size` - Image Size Limits

Checks total image size.

**Config:**
- `max_mb` (int) - hard limit in megabytes (fails if exceeded)
- `warn_mb` (int) - warning threshold (passes but warns)

**Example:**
```yaml
- id: "size-check"
  name: "Keep images lean"
  severity: "medium"
  kind: "size"
  config:
    max_mb: 2048
    warn_mb: 1024
```

### `label` - Required Labels

Ensures specific labels exist.

**Config:**
- `required` ([]string) - list of required label keys

**Example:**
```yaml
- id: "labels"
  name: "Require metadata labels"
  severity: "low"
  kind: "label"
  config:
    required:
      - "maintainer"
      - "version"
      - "build-date"
```

### `env` - Environment Variable Checks

Scans env vars for suspicious patterns (like hardcoded secrets).

**Config:**
- `deny_patterns` ([]string) - regex patterns that shouldn't appear

**Example:**
```yaml
- id: "no-secrets"
  name: "No secrets in ENV"
  severity: "critical"
  kind: "env"
  config:
    deny_patterns:
      - "(?i)password"
      - "(?i)api[_-]?key"
      - "(?i)secret"
```

### `base` - Base Image Validation

Checks if the base image is from a trusted set.

**Config:**
- `allowed_prefixes` ([]string) - acceptable base image prefixes
- `allow_unknown` (bool) - whether to allow unrecognized bases

**Example:**
```yaml
- id: "trusted-base"
  name: "Use approved base images"
  severity: "medium"
  kind: "base"
  config:
    allowed_prefixes:
      - "alpine"
      - "debian"
      - "distroless"
    allow_unknown: false
```

### `layers` - Layer Count Limits

Checks number of layers (too many can indicate inefficient builds).

**Config:**
- `max_layers` (int) - hard limit
- `warn_layers` (int) - warning threshold

**Example:**
```yaml
- id: "layer-count"
  name: "Minimize layers"
  severity: "low"
  kind: "layers"
  config:
    max_layers: 30
    warn_layers: 20
```

## Severity Levels

From most to least severe:

- **critical** - Must fix immediately (e.g., hardcoded secrets)
- **high** - Important security issue (e.g., running as root)
- **medium** - Should fix (e.g., unknown base image)
- **low** - Nice to fix (e.g., missing labels)
- **info** - Informational only

## Using Policies

### Built-in policies

```bash
# uses embedded default policy
boltguard nginx:latest

# uses policies/default.yaml if it exists
boltguard nginx:latest
```

### Custom policies

```bash
boltguard -policy my-policy.yaml nginx:latest
boltguard -policy /etc/boltguard/strict.yaml app:v1.2.3
```

### Policy locations

BoltGuard looks for default policy in:
1. `-policy` flag if specified
2. `policies/default.yaml` (project root)
3. `/etc/boltguard/default.yaml`
4. `~/.config/boltguard/default.yaml`
5. Embedded default

## Writing Custom Policies

Start with one of the examples in `policies/`:

- `default.yaml` - Balanced for general use
- `strict.yaml` - High-security production environments
- `permissive.yaml` - Development/testing

Copy and modify to match your risk tolerance.

## Policy Tips

- Start permissive, tighten gradually
- Use `fail_fast: true` on critical rules to stop early
- Set appropriate `min_severity` in CI (maybe "medium")
- Don't check for things you can't enforce
- Keep descriptions clear for developers who see failures
