# Contributing to crAIzy

Thank you for your interest in contributing to crAIzy! This document provides guidelines and instructions for contributing.

## Development Setup

1. **Prerequisites**
   - Go 1.21 or higher
   - tmux 3.0 or higher
   - Make
   - Git

2. **Clone the repository**
   ```bash
   git clone https://github.com/TechnicallyShaun/crAIzy.git
   cd crAIzy
   ```

3. **Install dependencies**
   ```bash
   make deps
   ```

4. **Build the project**
   ```bash
   make build
   ```

5. **Run tests**
   ```bash
   make test
   ```

## Development Workflow

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, documented code
   - Follow Go best practices
   - Add tests for new functionality

3. **Test your changes**
   ```bash
   # Run tests
   make test
   
   # Run linter
   make lint
   
   # Format code
   make fmt
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add feature: description of your changes"
   ```

5. **Push and create a Pull Request**
   ```bash
   git push origin feature/your-feature-name
   ```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting
- Write meaningful commit messages
- Add comments for exported functions and types

## Testing

- All new features must include unit tests
- Maintain or improve code coverage
- Test both happy paths and error cases
- Use table-driven tests where appropriate

Example test structure:
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Feature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Feature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Guidelines

1. **Title**: Use a clear, descriptive title
2. **Description**: Explain what changes you made and why
3. **Tests**: Include test results
4. **Documentation**: Update README or docs if needed
5. **Size**: Keep PRs focused and reasonably sized

## Reporting Issues

When reporting issues, please include:

- crAIzy version (`craizy version`)
- Operating system and version
- tmux version (`tmux -V`)
- Go version (`go version`)
- Steps to reproduce
- Expected vs actual behavior
- Error messages or logs

## Feature Requests

We welcome feature requests! Please:

1. Check if the feature already exists or is planned
2. Clearly describe the feature and use case
3. Explain why it would be valuable
4. Provide examples if possible

## Questions?

If you have questions:

- Open an issue with the "question" label
- Check existing issues and discussions
- Review the README and documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
