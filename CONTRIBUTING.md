# Contributing to BoltGuard

Thanks for your interest in contributing!

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker (for testing image inspection)
- Make (optional but recommended)

### Setup

```bash
# clone the repo
git clone https://github.com/R00TKI11/boltguard.git
cd boltguard

# download dependencies
go mod download

# build it
make build

# or without make
go build -o boltguard ./cmd/boltguard

# run tests
make test
```

## Development Workflow

1. **Fork** the repo
2. **Create a branch** for your feature/fix
3. **Make your changes** with tests
4. **Run tests and linting**:
   ```bash
   make test
   make lint
   make fmt
   ```
5. **Commit** with clear messages
6. **Push** and open a PR

## Code Style

- Keep it simple and readable
- Comment non-obvious logic
- Use `gofmt` (run `make fmt`)
- Pass `golangci-lint` (run `make lint`)

## Adding New Rule Types

To add a new rule evaluator:

1. Create evaluator in `internal/rules/evaluators.go`:
   ```go
   type MyEvaluator struct{}

   func (e *MyEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
       // your logic here
   }
   ```

2. Register it in `internal/rules/engine.go`:
   ```go
   e.Register("mykind", &MyEvaluator{})
   ```

3. Document it in `docs/policy.md`

4. Add example to `policies/examples/`

5. Write tests

## Project Structure

```
cmd/boltguard/     - CLI entry point
internal/
  image/           - Image loading and inspection
  facts/           - Fact extraction from images
  policy/          - Policy parsing
  rules/           - Rule engine and evaluators
  report/          - Output formatting
policies/          - Example policies
docs/              - Documentation
```

## Testing

We prefer practical tests over 100% coverage:

- Test actual use cases, not every line
- Test error paths that matter
- Integration tests are valuable
- Don't mock unless necessary

```bash
# run all tests
make test

# with coverage
make test-coverage
```

## Pull Request Guidelines

- Keep PRs focused on a single change
- Write a clear description
- Link related issues
- Ensure CI passes
- Update docs if needed

## Questions?

Open an issue or start a discussion. We're friendly!
