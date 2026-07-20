# kubectl-peek

[![CI](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yaml/badge.svg)](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/pierinho13/kubectl-peek)](https://github.com/pierinho13/kubectl-peek/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierinho13/kubectl-peek)](https://goreportcard.com/report/github.com/pierinho13/kubectl-peek)
[![License](https://img.shields.io/github/license/pierinho13/kubectl-peek)](LICENSE)

`kubectl-peek` is a lightweight, client-side Kubernetes productivity CLI for exploring cluster context, opening isolated working shells, switching namespaces, and inspecting Secrets together with the resources that use, produce, or reference them.

It brings three everyday Kubernetes workflows into one fast interactive tool:

- **Inspect Secrets** and decode their values while discovering related workloads, operators, and custom resources.
- **Open isolated Kubernetes shells** for a selected context and namespace without modifying the original kubeconfig.
- **Select and persist namespaces** interactively for the active or explicitly selected context.

Typical questions and tasks include:

- What is stored in this Secret?
- Which workloads use it?
- Which operator or custom resource produced it?
- What could be affected if it changes or is deleted?
- Which context and namespace should this terminal session use?
- How can I work in another Kubernetes scope without changing my normal shell configuration?

The tool runs entirely on the client side using your existing kubeconfig. It does not install controllers, agents, CRDs, web interfaces, or other components in the cluster.

<img width="1000" height="583" alt="kubectl-peek-demo-small" src="https://github.com/user-attachments/assets/b13f3b4d-da5b-46ed-97bb-ba07d3001e61" />



## Features

### Kubernetes shell workflows

- Interactive context and namespace selection with `kubectl-peek shell`.
- Direct shell startup with `--context` and `--namespace`.
- Isolated shells backed by temporary flattened kubeconfigs.
- Visible context and namespace indicators in the temporary shell prompt.
- Colored context and namespace prompt segments on supported terminals.
- Early protection against nested isolated shells.
- Temporary kubeconfig cleanup after leaving the shell.
- Original kubeconfig remains unchanged.

### Namespace workflows

- Interactive namespace selection.
- Keyboard navigation, pagination, and live filtering.
- Persistent namespace updates for the selected kubeconfig context.
- Optional initial namespace filtering.
- Context and kubeconfig overrides.

### Secret inspection

- Interactive Secret selection.
- Optional Secret-name filtering from the command line.
- Decoded Secret values displayed directly in the terminal.
- Built-in relationship discovery for Kubernetes workloads and resources.
- Declarative custom-resource discovery through YAML rules.
- Support for `uses`, `produces`, and `references` relationships.
- Automatic API-version selection for configured custom resources.
- Wildcard array traversal in custom-resource field paths.

### General

- Native `kubectl` plugin usage.
- No cluster-side installation required.
- Release binaries for macOS, Linux, and Windows.

## Why `kubectl-peek`?

Kubernetes users often jump between contexts, namespaces, terminal sessions, and Secret-related troubleshooting. Those tasks normally require a mix of `kubectl config`, temporary environment variables, manual kubeconfig copies, shell prompt customization, and several resource searches.

`kubectl-peek` turns those workflows into focused commands:

```text
kubectl-peek
├── secret       Inspect Secrets and their relationships
├── namespace    Select and persist a namespace
└── shell        Open an isolated context-aware Kubernetes shell
```

The root command displays help so every major workflow is immediately visible:

```bash
kubectl-peek
```

The same commands work through the native plugin form:

```bash
kubectl peek
```

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
kubectl-peek --version
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

Show the available workflows:

```bash
kubectl-peek
```

or:

```bash
kubectl peek
```

Inspect Secrets in the current namespace:

```bash
kubectl-peek secret
```

Open an isolated shell after selecting a context and namespace:

```bash
kubectl-peek shell
```

Select and persist a namespace for the active context:

```bash
kubectl-peek namespace
```

The standalone and native plugin forms are equivalent:

```bash
kubectl-peek secret
kubectl peek secret

kubectl-peek shell
kubectl peek shell

kubectl-peek namespace
kubectl peek namespace
```

## Usage

### Command overview

```text
kubectl-peek [command]

Available Commands:
  secret      Inspect Kubernetes Secrets and their relationships
  namespace   Select the namespace for a Kubernetes context
  shell       Open an isolated shell for a Kubernetes context and namespace
```

Global flags inherited by the commands:

```text
--context string      Kubernetes context
--kubeconfig string   Path to the kubeconfig file
```

### Inspect Secrets in the current namespace

```bash
kubectl-peek secret
```

Equivalent plugin command:

```bash
kubectl peek secret
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

Pass a pattern after the `secret` command:

```bash
kubectl-peek secret database
```

Matching is case-insensitive.

Equivalent plugin command:

```bash
kubectl peek secret database
```

### Inspect Secrets in another namespace

```bash
kubectl-peek secret -n staging
```

or:

```bash
kubectl peek secret -n staging
```

Combine namespace selection with a name pattern:

```bash
kubectl-peek secret database -n staging
```

### Inspect Secrets through another context

```bash
kubectl-peek secret --context development-cluster
```

### Use a specific kubeconfig

```bash
kubectl-peek secret --kubeconfig ~/.kube/secondary-config
```

### Combine Secret options

```bash
kubectl-peek secret database \
  --context development-cluster \
  --namespace staging \
  --kubeconfig ~/.kube/secondary-config
```

### Load custom discovery rules

```bash
kubectl-peek secret --rules ./rules.yaml
```

Combine custom rules with other options:

```bash
kubectl-peek secret database \
  --namespace staging \
  --rules ./examples/rules-all.yaml
```

Aliases are also available:

```bash
kubectl-peek secrets
kubectl-peek sec
```

## Context and namespace selection and isolated shells

`kubectl-peek` treats Secret inspection, namespace management, and isolated Kubernetes shells as first-class workflows.

### Select and persist a namespace

Run:

```bash
kubectl-peek namespace
```

or use the shorter alias:

```bash
kubectl-peek ns
```

This opens an interactive, paginated namespace selector with keyboard navigation and filtering.

Example:

```text
Select a namespace
Use ↑/↓ to move, ←/→ to change page, and / to filter.

  default
  monitoring
> traefik
  flux-system
  kube-system

Page 1/4 · 37 namespaces
```

After selecting a namespace, `kubectl-peek` updates the default namespace of the active kubeconfig context:

```text
Context "staging" now uses namespace "traefik"
```

You can provide an initial namespace filter:

```bash
kubectl-peek namespace traefik
```

You can also target a specific context or kubeconfig:

```bash
kubectl-peek namespace --context staging
kubectl-peek namespace --kubeconfig ~/.kube/secondary-config
```

The equivalent native plugin commands are:

```bash
kubectl peek namespace
kubectl peek ns
```

### Open an isolated context-aware Kubernetes shell

`kubectl-peek shell` is designed for safely working across multiple clusters and namespaces from separate terminal sessions. It selects a Kubernetes context and then a namespace from that context:

```bash
kubectl-peek shell
```

The native plugin form behaves identically:

```bash
kubectl peek shell
```

The interactive flow is:

```text
Select a Kubernetes context
            │
            ▼
List namespaces available through that context
            │
            ▼
Select a namespace
            │
            ▼
Create a temporary kubeconfig and open an isolated shell
```

A Kubernetes context is selected instead of a cluster directly because a context identifies the cluster and the credentials used to access it. After the context is known, `kubectl-peek` connects through that context and lists its available namespaces.

You can skip either selector by passing its value directly.

Use a specific context and select only the namespace:

```bash
kubectl-peek shell --context operations
```

Select a context interactively and use a specific namespace:

```bash
kubectl-peek shell --namespace traefik
```

or with the short namespace flag:

```bash
kubectl-peek shell -n traefik
```

Open the shell without interactive selectors:

```bash
kubectl-peek shell \
  --context operations \
  --namespace traefik
```

A namespace supplied through `--namespace` is validated against the selected context before the shell starts.

The command also supports a separate kubeconfig:

```bash
kubectl-peek shell \
  --kubeconfig ~/.kube/secondary-config
```

The following forms are equivalent:

```bash
kubectl-peek shell
kubectl peek shell
```

When no flags are supplied, the original kubeconfig and its current context are not modified. The selected context and namespace are applied only to the temporary kubeconfig used by the child shell.

If an isolated shell is already active, another one is rejected immediately, before displaying the context or namespace selectors:

```text
an isolated kubectl-peek shell is already active
(context "operations", namespace "traefik");
run exit before opening another one
```

Run `exit` before opening a different isolated shell.

### Open a namespace-first isolated shell

<img width="1800" height="800" alt="kubectl-peek-multi-namespace-shell" src="https://github.com/user-attachments/assets/fae7e6dd-1a36-491b-a4fe-0cbe5eaed4ee" />







Use `--shell` to open a child shell scoped to the selected context and namespace without modifying the original kubeconfig:

```bash
kubectl-peek namespace --shell
```

You can combine it with an initial filter:

```bash
kubectl-peek namespace traefik --shell
```

Or select a context explicitly:

```bash
kubectl-peek namespace \
  --context staging \
  --shell
```

When the shell starts, `kubectl-peek` displays the active scope:

```text
┌─ kubectl-peek namespace shell
│ Context    staging
│ Namespace  traefik
└─ Type `exit` to return to the previous shell
```

The child-shell prompt keeps the active Kubernetes scope visible:

```text
[k8s:staging ns:traefik] piero@MacBook-Pro %
```

The context is displayed in cyan and the namespace in yellow on terminals that support standard ANSI colors.

Commands executed inside the shell use the temporary kubeconfig:

```bash
kubectl get pods
kubectl get services
helm list
flux get all
```

The temporary kubeconfig:

- contains the complete effective Kubernetes configuration
- preserves clusters, users, contexts, authentication and exec plugins
- embeds certificate and key data referenced by file paths
- changes only the selected context namespace
- is created with `0600` permissions
- is exposed only to the child shell through `KUBECONFIG`
- does not modify the original kubeconfig

The temporary directory is removed after leaving the shell normally:

```bash
exit
```

Nested isolated shells are blocked to avoid accidentally creating multiple shell layers. The check happens before opening an interactive selector for both `kubectl-peek shell` and `kubectl-peek namespace --shell`.

## Interactive controls

The Secret, namespace, and context selectors share the same interaction model:

```text
↑ / ↓     Move through results
j / k     Move down or up
← / →     Change page when no interactive filter is active
/         Start filtering the visible result list
Enter     Select the highlighted result
Esc       Leave filtering mode or cancel
Ctrl+C    Cancel
```

When the interactive filter is active, type any substring from the Secret, namespace, or context name. The visible result list updates immediately.

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
kubectl-peek secret --rules ./rules.yaml
```

### With `KUBECTL_PEEK_RULE_FILE`

For a persistent default, define:

```bash
export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"
```

After that, running:

```bash
kubectl-peek secret
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

kubectl-peek secret --rules ./temporary-rules.yaml
```

uses `./temporary-rules.yaml` for that execution.

## Complete example output

```text
Secret: garage-admin-token
Namespace: sandbox-3
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
kubectl-peek secret --rules ./examples/rules-all.yaml
```

Review and customize the rules before using them in production environments. Operators and internal CRDs may use different field paths depending on their versions and configuration.

## Common workflows

### Investigate an application Secret

```bash
kubectl-peek secret database -n production
```

Select the Secret, inspect its decoded values, and review the workloads or custom resources connected to it.

### Open an isolated production troubleshooting shell

```bash
kubectl-peek shell \
  --context production   --namespace payments
```

Commands such as `kubectl`, `helm`, and `flux` use the temporary kubeconfig only inside that shell.

### Work across multiple clusters in parallel

Open separate terminal windows and run:

```bash
kubectl-peek shell --context staging -n payments
kubectl-peek shell --context production -n payments
```

Each terminal keeps its own isolated Kubernetes scope visible in the prompt.

### Persist a namespace for the active context

```bash
kubectl-peek namespace monitoring
```

Select a matching namespace and update the namespace stored in the chosen kubeconfig context.

### Extend discovery for an internal CRD

Create a YAML rule that points to the field containing the Secret name, then run:

```bash
kubectl-peek secret --rules ./company-rules.yaml
```

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
kubectl-peek secret --rules "$HOME/.config/kubectl-peek/rules.yaml"
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

### Verify the active isolated shell

Inside an isolated shell opened with `kubectl-peek shell` or `kubectl-peek namespace --shell`, inspect the temporary configuration with:

```bash
echo "$KUBECONFIG"
kubectl config current-context
kubectl config view --minify -o jsonpath='{..namespace}{"\n"}'
```

The prompt should also show the selected context and namespace:

```text
[k8s:staging ns:traefik] piero@MacBook-Pro %
```

Run `exit` to return to the previous shell and remove the temporary files.

## Development

Run all tests:

```bash
go test ./...
```

Build the project:

```bash
go build -o kubectl-peek .
```

Show the local command help:

```bash
./kubectl-peek
```

Run Secret inspection locally:

```bash
./kubectl-peek secret
```

Test custom rules:

```bash
./kubectl-peek secret --rules ./examples/rules-all.yaml
```

Test the isolated shell workflow:

```bash
./kubectl-peek shell
./kubectl-peek shell --context operations
./kubectl-peek shell --context operations --namespace traefik
```

Test the `kubectl` plugin form by placing the binary in your `PATH`:

```bash
kubectl peek
kubectl peek secret
kubectl peek namespace
kubectl peek shell
```

## Roadmap

Potential future improvements include:

- optional Secret value masking
- explicit `--show-values` behavior
- JSON and YAML output
- non-interactive Secret inspection
- dependency graph output
- ConfigMap discovery
- cross-namespace relationship discovery
- recursive relationship inspection
- reusable community rule collections
- additional built-in resource finders
- richer isolated-shell status and integrations
- additional context and namespace productivity workflows

## License

Licensed under the MIT License. See [LICENSE](LICENSE).