# Contributing to warp-wg

## Getting Started

1. Fork and clone the repository
2. Create a branch: `git switch -c <branch-name>`
3. Make your changes, then push and open a pull request

## Development

Requires Go (stable) and [golangci-lint](https://golangci-lint.run/) v2.

```bash
go build ./...
go test -race ./...
golangci-lint run
```

## Pull Requests

- All commits must be **signed** ([GPG or SSH](https://docs.github.com/en/authentication/managing-commit-signature-verification))
- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- One concern per pull request

## Security

Report vulnerabilities via [SECURITY.md](SECURITY.md), not through issues.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
