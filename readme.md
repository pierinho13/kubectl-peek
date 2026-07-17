# kubectl-peek

`kubectl-peek` is a small and straightforward CLI for browsing Kubernetes Secrets, seeing where they are used, and displaying their decoded values directly in the terminal.

The tool is intentionally designed to be simple:

- no complex configuration
- no additional services
- no custom resources
- no web interface
- no cluster-side installation

It works with your existing kubeconfig, active Kubernetes context, and current namespace.

## Features

- Interactive Secret selection.
- Simple keyboard navigation and pagination.
- Interactive filtering by Secret name.
- Optional name filtering from the command line.
- Namespace, context, and kubeconfig overrides.
- Decoded Secret values displayed directly in the terminal.
- `Used by` discovery for supported Kubernetes workloads.
- Support for macOS, Linux, and Windows release binaries.
- Native `kubectl` plugin usage.
- No cluster-side components required.

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

Both commands run the same binary and provide exactly the same functionality.

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

## Used by

When a Secret is selected, `kubectl-peek` searches the current namespace for supported workloads that reference it.

The current implementation detects usage in:

- Pods
- Deployments
- StatefulSets
- DaemonSets
- Jobs
- CronJobs

It detects references through:

- `env[].valueFrom.secretKeyRef`
- `envFrom[].secretRef`
- Secret-backed volumes
- `imagePullSecrets`
- Init containers
- Ephemeral containers

This makes it possible to inspect the Secret and understand its immediate workload dependencies in a single command.

Example:

```text
Secret: database-credentials
Namespace: default
Type: Opaque
Used by:
  Deployment/backend
    container/backend envFrom
  CronJob/backup
    container/backup env/BACKUP_PASSWORD -> password
```

When no supported workload references the Secret:

```text
Used by:
  none
```

The `Used by` result is limited to the supported workload types above. Other Kubernetes resources and custom resources may also reference Secrets.

## Example output

After selecting a Secret, `kubectl-peek` displays its metadata, detected usages, and decoded values:

```text
Secret: nebula-service-config
Namespace: default
Type: Opaque
Used by:
  Deployment/nebula-service
    container/nebula-service envFrom
  CronJob/nebula-backup
    container/backup env/BACKUP_PASSWORD -> password

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

The current Kubernetes identity must be allowed to read Secrets and list the supported workload resources in the selected namespace.

For Secret access:

```yaml
apiGroups:
  - ""
resources:
  - secrets
verbs:
  - get
  - list
```

For `Used by` discovery:

```yaml
apiGroups:
  - ""
resources:
  - pods
verbs:
  - list
---
apiGroups:
  - apps
resources:
  - deployments
  - statefulsets
  - daemonsets
verbs:
  - list
---
apiGroups:
  - batch
resources:
  - jobs
  - cronjobs
verbs:
  - list
```

You can verify Secret access with:

```bash
kubectl auth can-i list secrets -n default
kubectl auth can-i get secrets -n default
```

You can verify workload access with:

```bash
kubectl auth can-i list pods -n default
kubectl auth can-i list deployments.apps -n default
kubectl auth can-i list statefulsets.apps -n default
kubectl auth can-i list daemonsets.apps -n default
kubectl auth can-i list jobs.batch -n default
kubectl auth can-i list cronjobs.batch -n default
```

## Security notice

`kubectl-peek` prints decoded Secret values directly to the terminal.

Those values may remain visible in:

- terminal scrollback
- shell session recordings
- screen sharing sessions
- screenshots
- captured command output

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

You can also test it through the `kubectl` plugin form by placing the binary in your `PATH`:

```bash
kubectl peek
```

## License

Licensed under the MIT License. See [LICENSE](LICENSE).