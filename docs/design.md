# BoltGuard Design

## Philosophy

BoltGuard is built around a few core principles:

1. **Fast** - Should run in milliseconds, not seconds
2. **Offline-first** - Works in air-gapped environments with no registry access
3. **Policy-driven** - Human-readable YAML policies, not hardcoded checks
4. **Practical** - Focuses on checks that matter, not academic perfection

## Design constraints

BoltGuard is built around a small set of non-negotiable constraints:

- Offline-first operation by default
- No hidden network calls during scanning
- Deterministic and reproducible results
- Single static binary distribution
- Human-readable policy definitions

Features that violate these constraints are intentionally excluded or made
explicitly optional.

## Architecture

```
┌─────────────┐
│   CLI       │  Entry point, arg parsing
└──────┬──────┘
       │
       v
┌─────────────┐
│   Image     │  Load container image (daemon/tarball)
│   Loader    │  Extract config, manifest, layers
└──────┬──────┘
       │
       v
┌─────────────┐
│   Facts     │  Extract security-relevant metadata:
│   Extractor │  - user/root status
└──────┬──────┘  - labels, env vars
       │          - size, layers
       │          - base image
       v
┌─────────────┐
│   Policy    │  Parse YAML policy
│   Parser    │  Validate rules
└──────┬──────┘
       │
       v
┌─────────────┐
│   Rules     │  Evaluate each rule against facts
│   Engine    │  Return pass/fail + messages
└──────┬──────┘
       │
       v
┌─────────────┐
│   Report    │  Format results as:
│   Generator │  - text (human)
└─────────────┘  - json (CI)
                 - sarif (tooling)
```

## Design Decisions

### Why offline-first?

Many security tools require network access to vulnerability databases. This creates problems:
- Air-gapped environments can't use them
- Network flakes cause intermittent failures
- DBs get outdated quickly anyway
- External dependency adds clunkiness and complexity

BoltGuard focuses on policy checks that don't need external data.

### Why not scan for CVEs?

Vulnerability scanning is important but it's a different problem:
- Requires constantly updated databases
- Needs deep filesystem inspection
- Already solved well by other tools

BoltGuard complements CVE scanners with fast policy checks.

### Rule architecture

Each rule type implements the `Evaluator` interface:

```go
type Evaluator interface {
    Evaluate(facts *Facts, rule *policy.Rule) (*Result, error)
}
```

This makes it easy to add custom rule types without changing core logic.

### Future extensibility

v0.2 ideas:
- Offline advisory packs (curated policies for common frameworks)
- Layer file extraction (check for setuid binaries, etc.)
- Custom evaluator plugins (via Go plugins or WASM)
- Policy inheritance/composition

## Non-goals

Things BoltGuard deliberately doesn't do:

- **CVE scanning** - Use Trivy/Grype for that
- **Runtime protection** - Use Falco/AppArmor
- **Image signing** - Use Sigstore/Notary
- **Registry mirroring** - Use Harbor/distribution

BoltGuard does one thing: fast, offline policy checks.

### Why BoltGuard does not focus on CVE databases

CVE scanning is an important part of container security, but it introduces
trade-offs that conflict with BoltGuard’s goals:

- Large, frequently updated databases
- Increased binary size and startup time
- Reduced usefulness in air-gapped environments
- Less deterministic results across environments

BoltGuard intentionally focuses on **policy enforcement**, not vulnerability
enumeration. It is designed to complement, not replace, dedicated CVE scanners.

