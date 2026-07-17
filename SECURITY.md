# Security Policy

## Supported versions

Security fixes are applied to the latest released version of `kubectl-peek`.

## Reporting a vulnerability

Please do not report security vulnerabilities through public GitHub issues.

Use GitHub's private vulnerability reporting feature for this repository when available.

Include:

- a clear description of the issue
- affected versions
- steps to reproduce
- potential impact
- any suggested mitigation

Please avoid including real Kubernetes Secret values, credentials, access tokens, kubeconfig files, private cluster information, or other sensitive data.

## Security considerations

`kubectl-peek` reads Kubernetes Secrets and prints decoded values directly to the terminal.

Users should be aware that Secret values may remain visible in:

- terminal scrollback
- shell session recordings
- screenshots
- shared terminals
- screen-sharing sessions
- captured command output

Use the tool only in trusted environments and with the minimum Kubernetes permissions required.
