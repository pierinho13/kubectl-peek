# kubectl-peek

`kubectl-peek` is a small interactive CLI for browsing Kubernetes Secrets and displaying their decoded values directly from the terminal.

It uses the active Kubernetes context and namespace by default, while also supporting explicit namespace, context, and kubeconfig selection.

> Warning: selected Secret values are printed in plain text. Avoid using the tool while screen sharing or in terminals whose output is being recorded.

## Features

- Lists Secrets from the selected Kubernetes namespace.
- Interactive keyboard-based Secret selection.
- Pagination for large Secret lists.
- Interactive filtering by Secret name.
- Optional initial name filter from the command line.
- Supports custom namespaces, Kubernetes contexts, and kubeconfig files.
- Displays the selected Secret name, namespace, type, keys, and decoded values.
- Responsive filtered results that remain usable on smaller terminal windows.

## Requirements

- Access to a Kubernetes cluster.
- A valid kubeconfig.
- Permission to list and read Secrets in the target namespace.

The Kubernetes permissions required are equivalent to:

```yaml
apiGroups:
  - ""
resources:
  - secrets
verbs:
  - get
  - list
```

## Installation

Build the binary from the repository:

```bash
go build -o kubectl-peek .
```

Move it to a directory available in your `PATH`:

```bash
mv kubectl-peek /usr/local/bin/kubectl-peek
```

Alternatively, install it for the current user:

```bash
mkdir -p "$HOME/.local/bin"
mv kubectl-peek "$HOME/.local/bin/kubectl-peek"
```

Make sure the directory is included in your `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

Open the interactive Secret selector using the current Kubernetes context and namespace:

```bash
kubectl-peek
```

Use a specific namespace:

```bash
kubectl-peek --namespace monitoring
```

The namespace flag also supports its short form:

```bash
kubectl-peek -n monitoring
```

Use a specific Kubernetes context:

```bash
kubectl-peek --context staging
```

Use a custom kubeconfig file:

```bash
kubectl-peek --kubeconfig ~/.kube/staging-config
```

Combine the options:

```bash
kubectl-peek \
  --context staging \
  --namespace monitoring \
  --kubeconfig ~/.kube/config
```

## Filter Secrets from the command line

Pass a partial Secret name as the first argument:

```bash
kubectl-peek db
```

The match is case-insensitive and checks whether the Secret name contains the provided pattern.

For example, the previous command may show:

```text
Select a Secret from namespace "default"
Use ↑/↓ to move, ←/→ to change page, and / to filter.

> comet-db-credentials
  nebula-db-users
  driftwood-db
  harbor-db-admin
  orchard-db-credentials

Page 1/2 · 8 Secrets
```

You can combine the pattern with other options:

```bash
kubectl-peek db -n production
```

## Interactive controls

```text
↑ / ↓       Move through Secrets
← / →       Change page when no interactive filter is active
/           Start filtering the visible Secret list
Backspace   Remove characters from the filter
Enter       Select the highlighted Secret
Esc         Exit filtering or cancel the selector
Ctrl+C      Cancel the selector
```

When an interactive filter is active, matching results are displayed in a scrollable window rather than split into pages.

Example:

```text
Select a Secret from namespace "production"
/secret
Use ↑/↓ to move, ←/→ to change page, and / to filter.

  lantern-secret
  orchard-secret
> falcon-secret
  meadow-secret
  rocket-secret

18 matching Secrets · 3/18 selected
```

## Output

After selecting a Secret, `kubectl-peek` prints its metadata and all values contained in `data`.

Example output:

```text
Secret: harbor-db-admin
Namespace: production
Type: Opaque

password:
demo-password-123

username:
admin-user
```

Kubernetes represents Secret values as Base64 in YAML and JSON responses. The Kubernetes Go client exposes the `data` values as decoded byte arrays, so `kubectl-peek` prints their decoded contents directly.

## Examples

Browse Secrets in the current namespace:

```bash
kubectl-peek
```

Browse Secrets in `monitoring`:

```bash
kubectl-peek -n monitoring
```

Show only Secrets whose names contain `certificate`:

```bash
kubectl-peek certificate -n monitoring
```

Use the `cluster-eu` context:

```bash
kubectl-peek --context cluster-eu
```

Use another kubeconfig and namespace:

```bash
kubectl-peek token \
  --kubeconfig ~/.kube/cluster-config \
  --namespace sandbox
```

## Development

Run the application locally:

```bash
go run .
```

Format the code:

```bash
gofmt -w .
```

Run the tests:

```bash
go test ./...
```

Run the UI tests with verbose output:

```bash
go test -v ./internal/ui
```

Build the binary:

```bash
go build -o kubectl-peek .
```

## Current scope

The current version focuses only on Kubernetes Secrets. It can list, filter, select, and display decoded Secret values.

It does not currently modify Secrets, edit values, copy individual keys to the clipboard, inspect Secret usage, or integrate with external secret providers.

## License

Licensed under the MIT License. See [LICENSE](LICENSE).
