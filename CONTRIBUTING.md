# Contributing to kubectl-peek

Thank you for your interest in contributing to `kubectl-peek`.

The project aims to remain small, simple, and easy to use. Contributions should preserve that focus.

## Requirements

- Go installed
- A valid Kubernetes kubeconfig
- Access to a Kubernetes cluster for manual testing

## Development setup

Clone the repository:

```bash
git clone https://github.com/pierinho13/kubectl-peek.git
cd kubectl-peek
```

Download dependencies:

```bash
go mod download
```

Run the tests:

```bash
go test ./...
```

Build the project:

```bash
go build -o kubectl-peek .
```

Run it locally:

```bash
./kubectl-peek
```

## Code quality

Before opening a pull request, run:

```bash
gofmt -w .
go vet ./...
go test ./...
```

You can also validate the release configuration with:

```bash
goreleaser release --snapshot --clean
```

## Branches

Create a branch from `main`:

```bash
git checkout main
git pull
git checkout -b feat/my-change
```

Use a clear branch name, such as:

```text
feat/add-new-workload
fix/empty-secret-output
docs/update-installation
```

## Commits

Use clear and focused commit messages.

Examples:

```text
feat: add Secret usage discovery
fix: handle empty namespaces
docs: update Homebrew installation
test: cover Secret volume references
```

## Pull requests

A pull request should:

- explain the problem or feature
- describe the changes
- include tests when appropriate
- keep unrelated changes out of the same PR
- pass `go test ./...`
- preserve the tool's simple user experience

## Reporting bugs

Use the bug report template and include:

- `kubectl-peek` version
- operating system
- Kubernetes version
- relevant command
- expected behavior
- actual behavior
- steps to reproduce

Do not include real Secret values, tokens, credentials, kubeconfig contents, or other sensitive information.

## Feature requests

Feature requests are welcome, especially when they improve usability without adding unnecessary complexity.

Please describe:

- the problem you are trying to solve
- the expected behavior
- why it fits the scope of `kubectl-peek`

## Security issues

Do not report security vulnerabilities in public issues.

See [SECURITY.md](SECURITY.md) for reporting instructions.
