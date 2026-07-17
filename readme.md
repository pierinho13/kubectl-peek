# kubectl-peek

`kubectl-peek` is an interactive CLI for browsing Kubernetes Secrets and displaying their decoded values directly in the terminal.

It uses the active Kubernetes context and namespace by default, while also supporting explicit context, namespace, and kubeconfig selection.

## Features

- Interactive Secret selection.
- Keyboard navigation and pagination.
- Interactive filtering by Secret name.
- Optional name filtering from the command line.
- Namespace, context, and kubeconfig overrides.
- Decoded Secret values displayed in a readable format.
- Support for macOS, Linux, and Windows release binaries.
- Native `kubectl` plugin usage.

## Installation

### Homebrew

Add the tap and install `kubectl-peek`:

```bash
brew tap pierinho13/tools
brew install --cask kubectl-peek
```

You can also install it with the fully qualified tap name:

```bash
brew install --cask pierinho13/tools/kubectl-peek
```

Upgrade to the latest available version with:

```bash
brew update
brew upgrade --cask kubectl-peek
```

### GitHub Releases

Precompiled binaries for macOS, Linux, and Windows are published on the GitHub Releases page.

Download the archive for your operating system and architecture, extract it, and place the `kubectl-peek` binary in a directory included in your `PATH`.

Example for macOS or Linux:

```bash
tar -xzf kubectl-peek_<version>_<os>_<arch>.tar.gz
chmod +x kubectl-peek
sudo mv kubectl-peek /usr/local/bin/
```

Verify the installation:

```bash
kubectl-peek --help
```

### Build from source

Requirements:

- Go installed.
- Access to a Kubernetes cluster through a valid kubeconfig.

```bash
git clone https://github.com/pierinho13/kubectl-peek.git
cd kubectl-peek
go build -o kubectl-peek .
```

Move the binary into your `PATH`:

```bash
sudo mv kubectl-peek /usr/local/bin/
```

## Usage

Once installed, the tool can be executed in either of these forms:

```bash
kubectl-peek
```

or as a native `kubectl` plugin:

```bash
kubectl peek
```

Both commands run the same binary and provide the same functionality.

### Browse Secrets in the current namespace

```bash
kubectl-peek
```

Equivalent command:

```bash
kubectl peek
```

Example interactive view:

```text
Select a Secret from namespace "default"
Use ↑/↓ to move, ←/→ to change page, and / to filter.

  atlas-db-credentials
  comet-api-token
> nebula-service-config
  orchard-cache-password
  quartz-worker-auth

Page 1/3 · 24 Secrets
```

Press `Enter` to select the highlighted Secret.

### Filter by name from the command line

Pass a pattern as the first argument:

```bash
kubectl-peek database
```

Equivalent command:

```bash
kubectl peek database
```

Only Secrets whose names contain the supplied pattern are shown. Matching is case-insensitive.

Example:

```text
Select a Secret from namespace "default"
Use ↑/↓ to move, ←/→ to change page, and / to filter.

> atlas-database-password
  nebula-database-credentials
  quartz-database-token

Page 1/1 · 3 Secrets
```

### Use another namespace

```bash
kubectl-peek -n staging
```

or:

```bash
kubectl peek -n staging
```

You can combine the namespace with a name pattern:

```bash
kubectl-peek database -n staging
```

### Use another Kubernetes context

```bash
kubectl-peek --context development-cluster
```

### Use a specific kubeconfig

```bash
kubectl-peek --kubeconfig ~/.kube/secondary-config
```

Options can be combined:

```bash
kubectl-peek database \
  --context development-cluster \
  --namespace staging \
  --kubeconfig ~/.kube/secondary-config
```

## Interactive controls

```text
↑ / ↓     Move through Secrets
← / →     Change page when no interactive filter is active
/         Start filtering the visible Secret list
Enter     Select the highlighted Secret
Esc       Leave filtering mode or cancel
Ctrl+C    Cancel
```

When the interactive filter is active, type any substring of the Secret name. The result list updates immediately.

## Example output

After selecting a Secret, `kubectl-peek` displays its metadata and decoded values:

```text
Secret: nebula-service-config
Namespace: default
Type: Opaque

environment:
────────────────────────────────────────────────────────────
production

password:
────────────────────────────────────────────────────────────
example-password

username:
────────────────────────────────────────────────────────────
nebula-service
```

Kubernetes Secret values are returned by `client-go` as decoded byte values, so the application displays the decoded content directly.

## Permissions

The current Kubernetes identity must be allowed to list and read Secrets in the selected namespace.

At minimum, it needs permissions equivalent to:

```yaml
apiGroups:
  - ""
resources:
  - secrets
verbs:
  - get
  - list
```

You can verify access with:

```bash
kubectl auth can-i list secrets -n default
kubectl auth can-i get secrets -n default
```

## Security notice

`kubectl-peek` prints decoded Secret values directly to the terminal. Those values may remain visible in terminal scrollback, screen recordings, shared sessions, or captured command output.

Use the tool only in trusted environments and avoid exposing sensitive values unnecessarily.

## Development

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

## License

Licensed under the MIT License. See [LICENSE](LICENSE).
