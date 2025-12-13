# Contributing to Terraform Provider for Kinsta

Thank you for your interest in contributing to this provider! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Issues

Before creating an issue, please:

1. Search existing issues to avoid duplicates
2. Collect relevant information:
   - Provider version
   - Terraform/OpenTofu version
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant configuration files (sanitized)

### Suggesting Enhancements

Enhancement suggestions are welcome! Please:

1. Check if the enhancement is already proposed
2. Provide a clear use case
3. Explain the expected behavior
4. Consider backward compatibility

### Pull Requests

#### Before You Start

1. Open an issue first to discuss significant changes
2. Fork the repository
3. Create a feature branch from `main`

#### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/terraform-provider-kinsta
cd terraform-provider-kinsta

# Install dependencies
go mod download

# Build the provider
go build -o terraform-provider-kinsta
```

#### Making Changes

1. **Code Style**
   - Follow standard Go conventions
   - Run `go fmt` before committing
   - Use meaningful variable and function names
   - Add comments for complex logic

2. **Testing**
   - Write unit tests for new functionality
   - Ensure existing tests pass: `go test ./...`
   - For acceptance tests: `TF_ACC=true go test ./...`
   - Test coverage should not decrease

3. **Documentation**
   - Update relevant documentation in `docs/`
   - Add examples for new resources/data sources
   - Update README.md if needed
   - Document any breaking changes

4. **Commits**
   - Use clear, descriptive commit messages
   - Follow conventional commits format:
     ```
     type(scope): brief description

     Longer description if needed

     Breaking changes noted here
     ```
   - Types: feat, fix, docs, test, refactor, chore
   - Example: `feat(database): add support for custom backup retention`

#### Testing Your Changes

```bash
# Run unit tests
go test ./... -v

# Run specific package tests
go test ./internal/provider -v

# Run acceptance tests (requires API credentials)
export KINSTA_API_KEY=your_key
export KINSTA_COMPANY_ID=your_id
TF_ACC=true go test ./... -v

# Run linters (if configured)
golangci-lint run
```

#### Submitting a Pull Request

1. Push your changes to your fork
2. Create a pull request against `main`
3. Fill out the PR template completely
4. Link related issues using keywords (Fixes #123)
5. Wait for CI checks to pass
6. Address review feedback promptly

### PR Requirements

- [ ] All tests pass
- [ ] Code follows project style guidelines
- [ ] Documentation is updated
- [ ] Commits are well-formed
- [ ] No unnecessary dependencies added
- [ ] Breaking changes are clearly documented
- [ ] Examples are provided for new features

## Development Guidelines

### Code Organization

```
terraform-provider-kinsta/
├── internal/
│   ├── provider/       # Provider and resource implementations
│   └── client/         # API client code
├── docs/               # Provider documentation
├── examples/           # Usage examples
└── tests/              # Additional test files
```

### Resource Implementation

When adding a new resource:

1. Define the resource schema in `internal/provider/`
2. Implement CRUD operations
3. Add unit tests with mocks
4. Add acceptance tests
5. Generate documentation
6. Add example usage

### API Client

- Keep API interactions in `internal/client/`
- Add proper error handling
- Include retry logic for transient failures
- Use context for timeout/cancellation support

### Documentation

Provider documentation is generated from:
- Schema descriptions
- Markdown files in `docs/`
- Examples in resource files

Use `terraform-plugin-docs` (when set up) to generate registry-compatible docs.

## Testing Strategy

### Unit Tests

- Mock external dependencies
- Test edge cases and error conditions
- Use table-driven tests where appropriate
- Aim for >80% coverage of critical paths

### Acceptance Tests

- Test against real API (in isolated environment)
- Clean up resources after tests
- Use unique resource names to avoid conflicts
- Mark tests requiring API access appropriately

### Example Test Structure

```go
func TestResourceWordPressSite_Create(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

## Release Process

Releases are managed by maintainers:

1. Version bump following semver
2. Update CHANGELOG.md
3. Create and push version tag
4. GitHub Actions builds and publishes release
5. Registry updates automatically (once set up)

## Getting Help

- **Questions:** Open a GitHub Discussion
- **Bugs:** Create an issue with the bug template
- **Chat:** (Coming soon)

## Recognition

Contributors will be recognized in:
- Release notes
- CONTRIBUTORS file (if created)
- GitHub's contributor graph

## Legal

By contributing, you agree that your contributions will be licensed under the Mozilla Public License 2.0 (same as the project).

---

Thank you for contributing to make this provider better! 🎉
