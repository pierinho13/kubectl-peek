# kubectl-peek

[![CI](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yaml/badge.svg)](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/pierinho13/kubectl-peek)](https://github.com/pierinho13/kubectl-peek/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierinho13/kubectl-peek)](https://goreportcard.com/report/github.com/pierinho13/kubectl-peek)
[![License](https://img.shields.io/github/license/pierinho13/kubectl-peek)](LICENSE)

`kubectl-peek` is a lightweight, client-side Kubernetes productivity CLI for inspecting Secrets and their relationships, browsing Kubernetes Events, opening interactive Pod shells, switching namespaces, and creating isolated context-aware terminal sessions.

It brings five everyday Kubernetes workflows into one fast interactive tool:

- **Inspect Secrets** interactively, optionally redact their values, and discover related workloads, operators, and custom resources only when requested.
- **Browse Kubernetes Events** with grouping, filtering, occurrence counts, detailed views, and resource drill-down.
- **Exec into Pods interactively** by selecting a Pod and, when needed, one of its containers.
- **Open isolated Kubernetes shells** for a selected context and namespace without modifying the original kubeconfig.
- **Select and persist namespaces** interactively for the active or explicitly selected context.

Typical questions and tasks include:

- What is stored in this Secret, and which resources use or produce it?
- Which Warning Events are happening now, and how long have they been recurring?
- Which resource kinds and objects account for the most Event occurrences?
- Which Pod and container should I open for interactive troubleshooting?
- Which context and namespace should this terminal session use?
- How can I work in another Kubernetes scope without changing my normal shell configuration?

The tool runs entirely on the client side using your existing kubeconfig. It does not install controllers, agents, CRDs, web interfaces, or other components in the cluster.


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

<img width="1800" height="800" alt="kubectl-peek-shell-workflows" src="https://github.com/user-attachments/assets/94bd4178-0ecf-4750-bc31-1e46e155200b" />

### Namespace workflows

- Interactive namespace selection.
- Keyboard navigation, pagination, and live filtering.
- Persistent namespace updates for the selected kubeconfig context.
- Optional initial namespace filtering.
- Context and kubeconfig overrides.

<img width="1800" height="800" alt="kubectl-peek-multi-namespace-shell" src="https://github.com/user-attachments/assets/9eb0f303-cc02-4cc2-b434-ead0ceb10c35" />

### Secret inspection

- Interactive Secret selection.
- Optional Secret-name filtering from the command line.
- Decoded Secret values displayed directly in the terminal by default.
- Optional value redaction through `--show-values=false`, preserving each key and its byte length.
- Relationship discovery is opt-in through `--show-usage`.
- Built-in relationship discovery for Kubernetes workloads and resources.
- Declarative custom-resource discovery through YAML rules.
- Support for `uses`, `produces`, and `references` relationships.
- Automatic API-version selection for configured custom resources.
- Wildcard array traversal in custom-resource field paths.


<img width="1200" height="700" alt="kubectl-peek-demo-reducido" src="https://github.com/user-attachments/assets/02ec34c5-f82a-443c-a3f1-a03a2c9d5489" />

### Kubernetes event inspection

- Interactive Kubernetes Event browser ordered by latest occurrence.
- Grouped repeated events with clear occurrence counts, first/last seen timestamps, filtering, pagination, and detailed event inspection.
- Warning-focused and non-normal views through `--warnings` and `--non-normal`.
- Optional raw Event-object view with `--no-group`.
- Hierarchical exploration with `--browse` or `--browse-by-kind`.
- Navigate through `Kind → Resource → Events → Detail`.
- Aggregated resource, event, and occurrence counts at each level.
- Compatible with namespace, all-namespace, warning, non-normal, and text filters.

<img width="1400" height="760" alt="kubectl-peek-events-inspection" src="https://github.com/user-attachments/assets/ca59b59a-ab7c-4140-8e50-10b3f38b0efd" />

### Kubernetes pod exec

- Interactive Pod selection with readiness, container count, and current phase information.
- Optional Pod-name filtering directly from the command line.
- Automatic container selection when the Pod has only one container.
- Interactive container selection for multi-container Pods.
- Opens an interactive shell directly inside the selected container.
- Automatically tries `/bin/bash` and falls back to `/bin/sh` when needed.
- Supports explicit namespace, context, kubeconfig, container, and shell overrides.
- Direct interactive access with `kubectl-peek exec`.
- Navigate through `Pod → Container → Interactive shell`.
- Run troubleshooting commands inside the container and return with `exit`.
- Compatible with namespace, context, kubeconfig, Pod-name, container, and shell filters.

<img width="1200" height="700" alt="kubectl-peek-exec" src="https://github.com/user-attachments/assets/1a46465c-e129-4a63-9200-0658fed6de94" />



### General

- Native `kubectl` plugin usage.
- No cluster-side installation required.
- Release binaries for macOS, Linux, and Windows.

## Why `kubectl-peek`?

Kubernetes users often jump between contexts, namespaces, terminal sessions, Pod troubleshooting, Event inspection, and Secret-related investigation. Those tasks normally require a mix of `kubectl config`, temporary environment variables, manual kubeconfig copies, shell prompt customization, repeated `kubectl get` and `kubectl describe` commands, and several resource searches.

`kubectl-peek` turns those workflows into focused commands:

```text
kubectl-peek
├── secret       Inspect Secrets and their relationships
├── namespace    Select and persist a namespace
├── shell        Open an isolated context-aware Kubernetes shell
├── exec         Open an interactive shell inside a selected Pod
└── events       Browse, filter, group, and inspect Kubernetes Events
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

Secret relationship discovery is enabled explicitly with `--show-usage`.

```bash
kubectl-peek secret --show-usage
```

When enabled, `kubectl-peek` combines two discovery mechanisms.

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

Inspect a Secret and discover related resources:

```bash
kubectl-peek secret --show-usage
```

Inspect a Secret while keeping values redacted:

```bash
kubectl-peek secret --show-values=false
```

Browse Warning Events:

```bash
kubectl-peek events --warnings
```

Browse Events by resource kind and resource:

```bash
kubectl-peek events --warnings --browse
```

Open an interactive shell inside a selected Pod:

```bash
kubectl-peek exec
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

kubectl-peek events --warnings
kubectl peek events --warnings

kubectl-peek exec
kubectl peek exec

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
  exec        Open an interactive shell in a Kubernetes Pod
  events      Browse Kubernetes Events ordered by last occurrence
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

By default, `kubectl-peek` displays decoded Secret values and does not perform relationship discovery. This keeps the common inspection path fast and avoids unnecessary API calls or permission warnings.

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

### Discover resources related to a Secret

Relationship discovery is disabled by default.

Enable it explicitly with:

```bash
kubectl-peek secret --show-usage
```

This searches supported built-in Kubernetes resources and configured custom-resource rules for relationships such as `uses`, `produces`, and `references`.

You can combine it with a name filter and namespace:

```bash
kubectl-peek secret database \
  --namespace staging \
  --show-usage
```

When no supported relationship is found, the output explains that the result is not a guarantee that the Secret is unused:

```text
Used by:
  No references were found among the supported built-in resources and configured usage rules.
  This does not guarantee that the Secret is unused; unsupported resources, external systems, or unconfigured custom resources may still reference it.
```

### Hide Secret values

Decoded Secret values are shown by default.

Redact them with:

```bash
kubectl-peek secret --show-values=false
```

The key names remain visible and each value is replaced with its byte length:

```text
password:
────────────────────────────────────────────────────────────
<redacted: 24 bytes>
```

Relationship discovery and value redaction can be combined:

```bash
kubectl-peek secret \
  --show-usage \
  --show-values=false
```

### Load custom discovery rules

Custom rules are only evaluated when relationship discovery is enabled:

```bash
kubectl-peek secret \
  --show-usage \
  --rules ./rules.yaml
```

Combine custom rules with other options:

```bash
kubectl-peek secret database \
  --namespace staging \
  --show-usage \
  --rules ./examples/rules-all.yaml
```

Aliases are also available:

```bash
kubectl-peek secrets
kubectl-peek sec
```

## Pod exec

Use `kubectl-peek exec` to select a Pod interactively and open a shell inside one of its containers:

```bash
kubectl-peek exec
```

The selector shows the Pod name, readiness, container count, and current phase. Completed and failed Pods are excluded.

The interactive flow is:

```text
Select a Pod
      │
      ▼
Select a container when the Pod has more than one
      │
      ▼
Open /bin/bash or fall back to /bin/sh
```

Filter Pods by name:

```bash
kubectl-peek exec api
```

Use another namespace:

```bash
kubectl-peek exec -n staging
```

Select a specific container or shell directly:

```bash
kubectl-peek exec api \
  --namespace staging \
  --container application \
  --shell /bin/sh
```

The `x` alias is also available:

```bash
kubectl-peek x
kubectl peek x
```

Exit the container shell with `exit`.

## Kubernetes Event inspection

Browse Events in the current namespace:

```bash
kubectl-peek events
```

Events are ordered by their latest occurrence and repeated records are grouped by default. The table shows the latest occurrence, Event type, reason, involved object, and total occurrences.

Press `Enter` to display the selected Event's exact occurrence count, first-seen and last-seen timestamps, source, namespace, object, and message.

Show only Warning Events:

```bash
kubectl-peek events --warnings
```

Show every non-Normal Event type:

```bash
kubectl-peek events --non-normal
```

Show raw Event objects without grouping:

```bash
kubectl-peek events --no-group
```

Search across Event fields:

```bash
kubectl-peek events backoff
```

Use another namespace or all namespaces:

```bash
kubectl-peek events -n staging
kubectl-peek events -A
```

### Browse Events by resource

Use `--browse` or `--browse-by-kind` for hierarchical exploration:

```bash
kubectl-peek events --warnings --browse
```

The navigation flow is:

```text
Kind → Resource → Events → Detail
```

The first screen aggregates resources, Events, and occurrences by Kubernetes kind. The second shows the resources belonging to the selected kind. The final Event list contains only records related to the selected resource.

Browse mode works together with the other Event filters:

```bash
kubectl-peek events --non-normal --browse
kubectl-peek events backoff --browse
kubectl-peek events -A --warnings --browse
kubectl-peek events --no-group --browse
```

Aliases are also available:

```bash
kubectl-peek event
kubectl-peek ev
```

## Context and namespace selection and isolated shells

`kubectl-peek` treats Secret inspection, Event browsing, Pod exec, namespace management, and isolated Kubernetes shells as first-class workflows.

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

The Secret, namespace, context, Pod, container, Event, kind, and resource selectors share the same interaction model:

```text
↑ / ↓     Move through results
j / k     Move down or up
← / →     Change page when no interactive filter is active
/         Start filtering the visible result list
Enter     Select the highlighted result
Esc       Leave filtering mode or cancel
Ctrl+C    Cancel
```

When the interactive filter is active, type a matching substring. The visible result list updates immediately. Event filtering searches across several Event fields, while the other selectors filter by the visible resource or object name.

## Built-in Secret discovery

With `--show-usage`, `kubectl-peek` searches supported resources in the selected namespace without requiring additional configuration.

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

When no supported resource or configured rule references the Secret:

```text
Used by:
  No references were found among the supported built-in resources and configured usage rules.
  This does not guarantee that the Secret is unused; unsupported resources, external systems, or unconfigured custom resources may still reference it.
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

Array fields can be traversed with `[*]`.

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
path: spec.backends[*].credentials.secretName
```

Nested arrays are also supported:

```yaml
path: spec.groups[*].targets[*].secretRef.name
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
kubectl-peek secret \
  --show-usage \
  --rules ./rules.yaml
```

### With `KUBECTL_PEEK_RULE_FILE`

For a persistent default, define:

```bash
export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"
```

After that, running:

```bash
kubectl-peek secret --show-usage
```

automatically loads that file for relationship discovery.

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

kubectl-peek secret --show-usage --rules ./temporary-rules.yaml
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
kubectl-peek secret \
  --show-usage \
  --rules ./examples/rules-all.yaml
```

Review and customize the rules before using them in production environments. Operators and internal CRDs may use different field paths depending on their versions and configuration.

## Common workflows

### Investigate an application Secret

```bash
kubectl-peek secret database \
  -n production \
  --show-usage
```

Select the Secret, inspect its decoded values, and review the supported workloads or custom resources connected to it.

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

### Investigate recurring Warning Events

```bash
kubectl-peek events --warnings
```

Select an Event to inspect its occurrence count, first-seen and last-seen timestamps, source, and full message.

### Find the noisiest resources by Event kind

```bash
kubectl-peek events --warnings --browse
```

Navigate through resource kind, resource, matching Events, and Event details.

### Open a shell in an application Pod

```bash
kubectl-peek exec application -n production
```

Select the Pod and container, run troubleshooting commands, and leave the container with `exit`.

### Persist a namespace for the active context

```bash
kubectl-peek namespace monitoring
```

Select a matching namespace and update the namespace stored in the chosen kubeconfig context.

### Extend discovery for an internal CRD

Create a YAML rule that points to the field containing the Secret name, then run:

```bash
kubectl-peek secret \
  --show-usage \
  --rules ./company-rules.yaml
```

## Permissions

The active Kubernetes identity must have the permissions required by the selected workflow. Basic Secret inspection only requires reading Secrets. Related-resource permissions are needed only when `--show-usage` is enabled. Event inspection requires listing Events, and Pod exec requires listing/getting Pods plus access to the `pods/exec` subresource.

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

### Event inspection access

```yaml
apiGroups:
  - ""
resources:
  - events
verbs:
  - list
```

### Pod exec access

```yaml
apiGroups:
  - ""
resources:
  - pods
verbs:
  - get
  - list
---
apiGroups:
  - ""
resources:
  - pods/exec
verbs:
  - create
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
kubectl auth can-i list events -n default
kubectl auth can-i list pods -n default
kubectl auth can-i create pods/exec -n default
```

Insufficient permissions may prevent some relationships from being discovered.

## Security notice

`kubectl-peek` prints decoded Secret values directly to the terminal by default. Use `--show-values=false` when values should remain hidden.

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

You can also pass it explicitly while enabling discovery:

```bash
kubectl-peek secret \
  --show-usage \
  --rules "$HOME/.config/kubectl-peek/rules.yaml"
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
kubectl auth can-i list events -n default
kubectl auth can-i list pods -n default
kubectl auth can-i create pods/exec -n default
```

### The Secret has no `Used by` section

The `Used by` section is only displayed when `--show-usage` is enabled:

```bash
kubectl-peek secret --show-usage
```

### No supported Secret relationships are found

Possible reasons include:

- no supported resource references it
- the referencing resource is in another namespace
- the reference path is not yet supported
- a custom rule is missing
- the active identity cannot list the referencing resource
- the Secret is used externally rather than through the Kubernetes API

The discovery result should be treated as helpful dependency information, not as a guarantee that deleting a Secret is safe.

### No Events are displayed

Check that Events exist in the selected namespace, the active identity can list them, and the selected filters are not excluding every record.

```bash
kubectl get events -n default
kubectl auth can-i list events -n default
```

### Pod exec does not open a shell

Check that the Pod is running, the selected container exists, the active identity can create `pods/exec`, and the requested shell exists inside the container.

```bash
kubectl get pods -n default
kubectl auth can-i create pods/exec -n default
kubectl exec -n default <pod> -- /bin/sh
```

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

Test Secret redaction and usage discovery:

```bash
./kubectl-peek secret --show-values=false
./kubectl-peek secret --show-usage
```

Test custom rules:

```bash
./kubectl-peek secret \
  --show-usage \
  --rules ./examples/rules-all.yaml
```

Test Event inspection and browse mode:

```bash
./kubectl-peek events --warnings
./kubectl-peek events --warnings --browse
./kubectl-peek events -A --non-normal
```

Test Pod exec:

```bash
./kubectl-peek exec
./kubectl-peek exec api -n staging
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
kubectl peek events --warnings
kubectl peek exec
kubectl peek namespace
kubectl peek shell
```

## Roadmap

Potential future improvements include:

- JSON and YAML output
- non-interactive Secret inspection
- dependency graph output
- ConfigMap discovery
- cross-namespace relationship discovery
- recursive relationship inspection
- reusable community rule collections
- additional built-in resource finders
- richer isolated-shell status and integrations
- additional Event sorting and export formats
- optional live Event refresh
- Pod log browsing and container diagnostics
- additional context and namespace productivity workflows

## License

Licensed under the MIT License. See [LICENSE](LICENSE).