# kubectl-peek

[![CI](https://github.com/pierinho13/kubectl-peek/actions/workflows/ci.yml/badge.svg)](https://github.com/pierinho13/kubectl-peek/actions)
[![Release](https://img.shields.io/github/v/release/pierinho13/kubectl-peek)](https://github.com/pierinho13/kubectl-peek/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/pierinho13/kubectl-peek)](https://goreportcard.com/report/github.com/pierinho13/kubectl-peek)
[![License](https://img.shields.io/github/license/pierinho13/kubectl-peek)](LICENSE)

**An interactive Kubernetes troubleshooting and terminal-workflow CLI.**

`kubectl-peek` helps you inspect Secret relationships, investigate Events, enter Pod containers, select namespaces, and work safely across multiple Kubernetes contexts without changing the kubeconfig used by your other terminal sessions.

```text
kubectl peek
├── secret       Inspect Secrets and discover related resources
├── events       Group, filter and investigate Kubernetes Events
├── exec         Select a Pod and open an interactive container shell
├── namespace    Select and persist a namespace
└── shell        Open an isolated context-aware Kubernetes shell
```

- Client-side only
- Uses your existing kubeconfig and Kubernetes permissions
- Installs no controller, agent, CRD or web interface
- Supports macOS, Linux and Windows
- Works as both `kubectl-peek` and the native `kubectl peek` plugin

## Why kubectl-peek?

Everyday Kubernetes troubleshooting often requires several disconnected commands: changing contexts, selecting namespaces, finding Pods and containers, reading repetitive Events, inspecting Secrets, and tracing the resources that use them.

`kubectl-peek` turns those workflows into focused interactive experiences while keeping the active Kubernetes scope visible and predictable.

| Workflow | Problem it solves | Command |
|---|---|---|
| Isolated shell | Work in another context without changing other terminals | `kubectl peek shell` |
| Events | Group noisy Events and drill into affected resources | `kubectl peek events --warnings --browse` |
| Secrets | Decode a Secret and discover who uses or produces it | `kubectl peek secret` |
| Pod access | Select the correct Pod and container interactively | `kubectl peek exec` |
| Namespace | Select and persist a namespace without remembering its name | `kubectl peek namespace` |

## Isolated Kubernetes shells

Open independent terminal sessions for different clusters and namespaces without modifying the original kubeconfig:

```bash
kubectl peek shell
```

![kubectl-peek shell workflows](https://github.com/user-attachments/assets/94bd4178-0ecf-4750-bc31-1e46e155200b)

Each child shell receives a temporary flattened kubeconfig through `KUBECONFIG`. The prompt keeps the selected context and namespace visible:

```text
[k8s:production ns:payments] user@host %
```

Tools such as `kubectl`, `helm` and `flux` use the isolated scope only inside that shell. Temporary files are removed after `exit`.

[Full shell and namespace documentation](docs/shells-and-namespaces.md)

## Namespace workflows

Select and persist a namespace for the active or explicitly selected context:

```bash
kubectl peek namespace
kubectl peek namespace monitoring
kubectl peek namespace --context staging
```

![kubectl-peek namespace workflow](https://github.com/user-attachments/assets/9eb0f303-cc02-4cc2-b434-ead0ceb10c35)

A namespace-first isolated shell is also available:

```bash
kubectl peek namespace --shell
```

![kubectl-peek namespace shell](https://github.com/user-attachments/assets/fae7e6dd-1a36-491b-a4fe-0cbe5eaed4ee)

[Full shell and namespace documentation](docs/shells-and-namespaces.md)

## Secret inspection and relationship discovery

Select a Secret interactively, decode its values and discover workloads, operators or custom resources related to it:

```bash
kubectl peek secret
kubectl peek secret database -n production
kubectl peek secret --rules ./examples/rules-all.yaml
```

![kubectl-peek Secret inspection](https://github.com/user-attachments/assets/02ec34c5-f82a-443c-a3f1-a03a2c9d5489)

Built-in discovery covers common Kubernetes resources. Declarative YAML rules extend discovery to resources such as External Secrets Operator, cert-manager, Crossplane, Vault operators, Argo CD and internal CRDs.

[Secret inspection](docs/secrets.md) · [Custom discovery rules](docs/rules.md)

## Kubernetes Event investigation

Browse Events ordered by their latest occurrence, group repeated records and inspect occurrence counts, first/last seen timestamps, sources and messages:

```bash
kubectl peek events
kubectl peek events --warnings
kubectl peek events --warnings --browse
kubectl peek events -A --non-normal
```

![kubectl-peek Event inspection](https://github.com/user-attachments/assets/ca59b59a-ab7c-4140-8e50-10b3f38b0efd)

Browse mode provides hierarchical navigation:

```text
Kind → Resource → Events → Detail
```

[Full Event documentation](docs/events.md)

## Interactive Pod exec

Select a running Pod, choose a container when necessary, and open an interactive shell:

```bash
kubectl peek exec
kubectl peek exec api -n staging
kubectl peek exec api --container application --shell /bin/sh
```

![kubectl-peek exec](https://github.com/user-attachments/assets/1a46465c-e129-4a63-9200-0658fed6de94)

`kubectl-peek` tries `/bin/bash` first and falls back to `/bin/sh`.

[Full Pod exec documentation](docs/exec.md)

## Installation

### Homebrew

```bash
brew tap pierinho13/tools
brew install --cask kubectl-peek
```

Upgrade:

```bash
brew update
brew upgrade --cask kubectl-peek
```

### GitHub Releases

Download the archive for your operating system and architecture from the [Releases page](https://github.com/pierinho13/kubectl-peek/releases), extract it, and place the binary in your `PATH`.

```bash
tar -xzf kubectl-peek_<version>_<os>_<arch>.tar.gz
chmod +x kubectl-peek
sudo mv kubectl-peek /usr/local/bin/
```

### Build from source

```bash
git clone https://github.com/pierinho13/kubectl-peek.git
cd kubectl-peek
go build -o kubectl-peek .
sudo mv kubectl-peek /usr/local/bin/
```

Verify:

```bash
kubectl peek --help
kubectl peek --version
```

## Quick start

```bash
# Show all workflows
kubectl peek

# Inspect a Secret
kubectl peek secret

# Investigate Warning Events
kubectl peek events --warnings --browse

# Open a container shell
kubectl peek exec

# Open an isolated Kubernetes shell
kubectl peek shell

# Persist a namespace
kubectl peek namespace
```

Global overrides:

```text
--context string      Kubernetes context
--kubeconfig string   Path to the kubeconfig file
```

## Interactive controls

```text
↑ / ↓     Move through results
j / k     Move down or up
← / →     Change page
/         Filter visible results
Enter     Select
Esc       Leave filter mode or cancel
Ctrl+C    Cancel
```

## Security model

`kubectl-peek` runs locally and uses the permissions of the active Kubernetes identity. It does not install anything in the cluster.

Secret values are decoded and printed to the terminal. They may remain visible in terminal scrollback, recordings, screen sharing, screenshots or captured output. Use Secret inspection only in trusted environments.

Temporary kubeconfigs created for isolated shells:

- contain the effective Kubernetes configuration;
- preserve authentication and exec plugins;
- embed referenced certificate and key data;
- are created with `0600` permissions;
- are exposed only to the child shell;
- are removed after the shell exits normally;
- do not modify the original kubeconfig.

[Permissions and security details](docs/permissions-and-security.md)

## Documentation

- [Shells and namespaces](docs/shells-and-namespaces.md)
- [Secret inspection](docs/secrets.md)
- [Custom discovery rules](docs/rules.md)
- [Event investigation](docs/events.md)
- [Pod exec](docs/exec.md)
- [Permissions and security](docs/permissions-and-security.md)
- [Development and troubleshooting](docs/development.md)

## Roadmap

Potential improvements include:

- optional Secret value masking and explicit reveal behavior;
- JSON and YAML output;
- non-interactive workflows;
- dependency graph output;
- ConfigMap and cross-namespace discovery;
- reusable community rule collections;
- live Event refresh and export formats;
- Pod log browsing and container diagnostics;
- richer isolated-shell status and integrations.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) and [SECURITY.md](SECURITY.md).

## License

Licensed under the [MIT License](LICENSE).
