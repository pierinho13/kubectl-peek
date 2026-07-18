# kubectl-peek

[![CI](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yaml/badge.svg)](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/pierinho13/kubectl-peek)](https://github.com/pierinho13/kubectl-peek/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierinho13/kubectl-peek)](https://goreportcard.com/report/github.com/pierinho13/kubectl-peek)
[![License](https://img.shields.io/github/license/pierinho13/kubectl-peek)](LICENSE)

`kubectl-peek` is a lightweight CLI for interactively browsing Kubernetes Secrets, inspecting their decoded values, and understanding which resources use, produce, or reference them.

Instead of only answering:

> What is stored in this Secret?

`kubectl-peek` also helps answer:

- Which workloads use this Secret?
- Which operator or custom resource produced it?
- Which resources reference it?
- What could be affected if the Secret changes or is deleted?

The tool runs entirely on the client side using your existing kubeconfig. It does not install controllers, agents, CRDs, web interfaces, or other components in the cluster.

<img width="1000" height="583" alt="kubectl-peek-demo-small" src="https://github.com/user-attachments/assets/b13f3b4d-da5b-46ed-97bb-ba07d3001e61" />



## Features

- Interactive Secret selection.
- Keyboard navigation, pagination, and interactive filtering.
- Optional Secret-name filtering from the command line.
- Namespace, context, and kubeconfig overrides.
- Decoded Secret values displayed directly in the terminal.
- Built-in Secret relationship discovery.
- Custom-resource discovery through declarative YAML rules.
- Support for `uses`, `produces`, and `references` relationships.
- Automatic API-version selection for configured custom resources.
- Support for wildcard paths in custom-resource rules.
- Native `kubectl` plugin usage.
- No cluster-side installation required.
- Release binaries for macOS, Linux, and Windows.

## How discovery works

`kubectl-peek` combines two discovery mechanisms.

### Built-in discovery

Common Kubernetes resources are inspected automatically. No rule file is required.

### Rule-based discovery

Custom resources and additional Kubernetes resource types can be supported through a YAML rules file.

This allows `kubectl-peek` to understand resources such as:

- External Secrets Operator `ExternalSecret`
- cert-manager `Certificate`
- Crossplane resources
- Vault operators
- Argo CD resources
- internal company CRDs
- any namespaced Kubernetes resource that stores a Secret name at a predictable field path

The application itself does not need hardcoded knowledge of every CRD.

```text
                         Secret
                            │
              ┌─────────────┴─────────────┐
              │                           │
       Built-in finders             YAML rules
              │                           │
       Deployments                  ExternalSecret
       StatefulSets                Certificate
       Pods                        Crossplane resources
       ServiceAccounts             Internal CRDs
       Ingresses                    Other operators
       Gateways
              │                           │
              └─────────────┬─────────────┘
                            │
                       Relationships
                  uses / produces / references
```

## Installation

### Homebrew

For macOS and Linux:

```bash
brew tap pierinho13/tools
brew install --cask kubectl-peek
```

You can also use the fully qualified tap name:

```bash
brew install --cask pierinho13/tools/kubectl-peek
```

Upgrade to the latest available version:

```bash
brew update
brew upgrade --cask kubectl-peek
```

### GitHub Releases

Precompiled binaries for macOS, Linux, and Windows are published on the GitHub Releases page.

Download the archive for your operating system and architecture, extract it, and place the binary in a directory included in your `PATH`.

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

- Go
- Access to a Kubernetes cluster through a valid kubeconfig

```bash
git clone https://github.com/pierinho13/kubectl-peek.git
cd kubectl-peek
go build -o kubectl-peek .
```

Move the binary into your `PATH`:

```bash
sudo mv kubectl-peek /usr/local/bin/
```

## Quick start

Run the standalone command:

```bash
kubectl-peek
```

Or use it as a native `kubectl` plugin:

```bash
kubectl peek
```

Both forms execute the same binary and provide the same functionality.

## Usage

### Browse Secrets in the current namespace

```bash
kubectl-peek
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

Press `Enter` to inspect the highlighted Secret.

### Filter by Secret name

Pass a pattern as the first positional argument:

```bash
kubectl-peek database
```

Matching is case-insensitive.

Equivalent plugin command:

```bash
kubectl peek database
```

### Use another namespace

```bash
kubectl-peek -n staging
```

or:

```bash
kubectl peek -n staging
```

You can combine namespace selection with a name pattern:

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

### Combine options

```bash
kubectl-peek database \
  --context development-cluster \
  --namespace staging \
  --kubeconfig ~/.kube/secondary-config
```

### Load custom discovery rules

```bash
kubectl-peek --rules ./rules.yaml
```

You can combine custom rules with the other options:

```bash
kubectl-peek database \
  --namespace staging \
  --rules ./examples/rules-all.yaml
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

When the interactive filter is active, type any substring from the Secret name. The visible result list updates immediately.

## Built-in Secret discovery

Without additional configuration, `kubectl-peek` searches supported resources in the selected namespace.

The built-in discovery currently includes common workload and Secret-related resources such as:

- Pods
- Deployments
- StatefulSets
- DaemonSets
- Jobs
- CronJobs
- ServiceAccounts
- Ingresses
- Gateway API Gateways, when the API is installed

The exact set may grow over time.

Built-in finders detect references such as:

- `env[].valueFrom.secretKeyRef`
- `envFrom[].secretRef`
- Secret-backed volumes
- projected Secret volumes
- `imagePullSecrets`
- init-container Secret references
- ephemeral-container Secret references
- ServiceAccount pull Secrets
- TLS Secret references
- other supported built-in paths

Example:

```text
Secret: database-credentials
Namespace: default
Type: Opaque
Used by:
  Deployment/backend
    uses: environment variable (container/backend env/DATABASE_PASSWORD -> password)
  CronJob/backup
    uses: environment variable (container/backup env/BACKUP_PASSWORD -> password)
```

When no supported resource references the Secret:

```text
Used by:
  none
```

## Custom resource discovery

Kubernetes clusters frequently rely on operators and CRDs that create or consume Secrets.

Hardcoding every possible operator in `kubectl-peek` would be difficult to maintain and would prevent users from supporting private CRDs.

Instead, custom relationships are described with a declarative YAML rules file.

### Example: External Secrets Operator

```yaml
rules:
  - apiVersions:
      - external-secrets.io/v1
      - external-secrets.io/v1beta1
    kind: ExternalSecret
    resource: externalsecrets
    references:
      - path: spec.target.name
        relation: produces
        description: generated Secret
```

This tells `kubectl-peek` that:

- the Kubernetes kind is `ExternalSecret`
- its plural API resource is `externalsecrets`
- either listed API version may be used
- the generated Secret name is stored in `spec.target.name`
- the relationship should be displayed as `produces`

Example result:

```text
ExternalSecret/garage-admin-token
  produces: generated Secret (spec.target.name)
```

### Example: cert-manager

```yaml
rules:
  - apiVersions:
      - cert-manager.io/v1
    kind: Certificate
    resource: certificates
    references:
      - path: spec.secretName
        relation: produces
        description: issued TLS Secret
```

Example result:

```text
Certificate/web-tls
  produces: issued TLS Secret (spec.secretName)
```

## Rule-file format

A rule file contains a top-level `rules` array:

```yaml
rules:
  - apiVersions:
      - example.io/v1
    kind: Example
    resource: examples
    references:
      - path: spec.secretName
        relation: uses
        description: application credentials
```

### Rule fields

| Field | Required | Description |
|---|---:|---|
| `apiVersions` | Yes | API versions that may provide this resource, in preferred order. |
| `kind` | Yes | Kubernetes kind shown in results. |
| `resource` | Yes | Plural API resource name used by the Kubernetes API. |
| `references` | Yes | One or more Secret-reference definitions. |

### Reference fields

| Field | Required | Description |
|---|---:|---|
| `path` | Yes | Dot-separated path that contains a Secret name. |
| `relation` | Recommended | Relationship type: `uses`, `produces`, or `references`. |
| `description` | Recommended | Human-readable explanation shown in the output. |

### Relationships

#### `uses`

Use when the resource consumes the Secret.

```yaml
relation: uses
```

Example output:

```text
Database/app
  uses: database credentials (spec.credentialsSecret.name)
```

#### `produces`

Use when the resource creates, synchronizes, or owns the resulting Secret.

```yaml
relation: produces
```

Example output:

```text
ExternalSecret/app
  produces: generated Secret (spec.target.name)
```

#### `references`

Use for a neutral relationship when neither `uses` nor `produces` is fully accurate.

```yaml
relation: references
```

Example output:

```text
Application/app
  references: configured Secret (spec.secretRef.name)
```

## Paths and wildcard arrays

Paths are dot-separated field paths relative to the Kubernetes object.

Simple example:

```yaml
path: spec.secretName
```

Nested example:

```yaml
path: spec.credentials.secretRef.name
```

Array fields can be traversed with `*`.

Example object:

```yaml
spec:
  backends:
    - name: primary
      credentials:
        secretName: database-primary
    - name: replica
      credentials:
        secretName: database-replica
```

Matching rule:

```yaml
path: spec.backends.*.credentials.secretName
```

Nested arrays are also supported:

```yaml
path: spec.groups.*.targets.*.secretRef.name
```

Use paths that resolve directly to the Secret name string.

## Multiple API versions

A rule can list multiple versions:

```yaml
apiVersions:
  - external-secrets.io/v1
  - external-secrets.io/v1beta1
```

`kubectl-peek` checks the cluster discovery API and uses the first available resource version.

This makes one rule file usable across clusters running different operator versions.

If none of the configured API versions is installed, the rule is skipped.

## Loading rule files

Custom rules are optional.

### With `--rules`

```bash
kubectl-peek --rules ./rules.yaml
```

### With `KUBECTL_PEEK_RULE_FILE`

For a persistent default, define:

```bash
export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"
```

After that, running:

```bash
kubectl-peek
```

automatically loads that file.

A common setup is:

```bash
mkdir -p "$HOME/.config/kubectl-peek"
cp examples/rules-all.yaml "$HOME/.config/kubectl-peek/rules.yaml"

echo 'export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"' \
  >> "$HOME/.zshrc"
```

Reload the shell:

```bash
source "$HOME/.zshrc"
```

### Precedence

The rule-file path is resolved in this order:

1. Explicit `--rules` flag
2. `KUBECTL_PEEK_RULE_FILE`
3. Built-in discovery only

For example:

```bash
export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"

kubectl-peek --rules ./temporary-rules.yaml
```

uses `./temporary-rules.yaml` for that execution.

## Complete example output

```text
Secret: garage-admin-token
Namespace: devbox-sre-3
Type: Opaque
Used by:
  ExternalSecret/garage-admin-token
    produces: generated Secret (spec.target.name)
  StatefulSet/garage
    uses: Secret volume (volume/admin-token)

admin-token:
────────────────────────────────────────────────────────────
example-value
```

Built-in and rule-based results are combined into the same `Used by` section.

## Example rule files

The repository includes reusable examples under [`examples/`](examples/):

- [`rules-external-secrets.yaml`](examples/rules-external-secrets.yaml)
- [`rules-cert-manager.yaml`](examples/rules-cert-manager.yaml)
- [`rules-crossplane.yaml`](examples/rules-crossplane.yaml)
- [`rules-common.yaml`](examples/rules-common.yaml)
- [`rules-all.yaml`](examples/rules-all.yaml)

Start with:

```bash
kubectl-peek --rules ./examples/rules-all.yaml
```

Review and customize the rules before using them in production environments. Operators and internal CRDs may use different field paths depending on their versions and configuration.

## Permissions

The active Kubernetes identity must be allowed to read Secrets and list the resources used for discovery.

### Secret access

```yaml
apiGroups:
  - ""
resources:
  - secrets
verbs:
  - get
  - list
```

### Common built-in discovery access

```yaml
apiGroups:
  - ""
resources:
  - pods
  - serviceaccounts
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
---
apiGroups:
  - networking.k8s.io
resources:
  - ingresses
verbs:
  - list
---
apiGroups:
  - gateway.networking.k8s.io
resources:
  - gateways
verbs:
  - list
```

Custom rules also require `list` permission for each configured resource.

Example for External Secrets Operator:

```yaml
apiGroups:
  - external-secrets.io
resources:
  - externalsecrets
verbs:
  - list
```

### Verify permissions

```bash
kubectl auth can-i list secrets -n default
kubectl auth can-i get secrets -n default
kubectl auth can-i list deployments.apps -n default
kubectl auth can-i list externalsecrets.external-secrets.io -n default
```

Insufficient permissions may prevent some relationships from being discovered.

## Security notice

`kubectl-peek` prints decoded Secret values directly to the terminal.

Those values may remain visible in:

- terminal scrollback
- shell session recordings
- shared terminal sessions
- screen-sharing sessions
- screenshots
- captured command output
- CI logs, if the command is used in automation

Use the tool only in trusted environments and avoid exposing sensitive values unnecessarily.

Do not paste real Secret values into public bug reports or pull requests.

## Troubleshooting

### The rule file is not loaded

Check the configured path:

```bash
echo "$KUBECTL_PEEK_RULE_FILE"
```

Then verify that the file exists:

```bash
ls -l "$KUBECTL_PEEK_RULE_FILE"
```

You can also pass it explicitly:

```bash
kubectl-peek --rules "$HOME/.config/kubectl-peek/rules.yaml"
```

### A custom resource is not detected

Check:

1. The configured `apiVersions` match an installed API version.
2. `resource` uses the plural API resource name, not the kind.
3. `path` resolves directly to the Secret name.
4. The resource exists in the selected namespace.
5. Your Kubernetes identity can list the resource.
6. The YAML file is valid.

Useful commands:

```bash
kubectl api-resources | grep -i external
kubectl get externalsecrets.external-secrets.io -n default
kubectl auth can-i list externalsecrets.external-secrets.io -n default
```

### The Secret has no `Used by` entries

Possible reasons include:

- no supported resource references it
- the referencing resource is in another namespace
- the reference path is not yet supported
- a custom rule is missing
- the active identity cannot list the referencing resource
- the Secret is used externally rather than through the Kubernetes API

The discovery result should be treated as helpful dependency information, not as a guarantee that deleting a Secret is safe.

## Development

Run all tests:

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

Test custom rules:

```bash
./kubectl-peek --rules ./examples/rules-all.yaml
```

Test the `kubectl` plugin form by placing the binary in your `PATH`:

```bash
kubectl peek
```

## Contributing rules

Contributions for commonly used operators are welcome.

A useful rule contribution should:

- support currently maintained API versions
- use the correct plural API resource name
- point directly to Secret-name fields
- use an accurate relationship
- include a clear description
- avoid speculative or version-specific paths unless documented

Before submitting a rule, verify it against real resource manifests and run:

```bash
go test ./...
```

## Roadmap

Potential future improvements include:

- optional value masking
- explicit `--show-values` behavior
- JSON and YAML output
- non-interactive mode
- dependency graph output
- ConfigMap discovery
- cross-namespace relationship discovery
- recursive relationship inspection
- reusable community rule collections
- additional built-in resource finders

## License

Licensed under the MIT License. See [LICENSE](LICENSE).
