# Custom Secret discovery rules

[← Back to README](../README.md)

Kubernetes clusters commonly use operators and internal CRDs that create or consume Secrets. `kubectl-peek` supports these resources through declarative YAML rules rather than hardcoding every operator.

## Example: External Secrets Operator

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

## Example: cert-manager

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

## Rule-file format

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
| `apiVersions` | Yes | API versions that may provide the resource, in preferred order |
| `kind` | Yes | Kubernetes kind displayed in results |
| `resource` | Yes | Plural API resource name |
| `references` | Yes | One or more Secret-reference definitions |

### Reference fields

| Field | Required | Description |
|---|---:|---|
| `path` | Yes | Dot-separated path containing the Secret name |
| `relation` | Recommended | `uses`, `produces` or `references` |
| `description` | Recommended | Human-readable explanation |

## Relationship types

Use `uses` when the resource consumes the Secret:

```yaml
relation: uses
```

Use `produces` when it creates, synchronizes or owns the Secret:

```yaml
relation: produces
```

Use `references` for a neutral relationship:

```yaml
relation: references
```

## Paths and wildcard arrays

Simple path:

```yaml
path: spec.secretName
```

Nested path:

```yaml
path: spec.credentials.secretRef.name
```

Array traversal:

```yaml
path: spec.backends[*].credentials.secretName
```

Nested arrays:

```yaml
path: spec.groups[*].targets[*].secretRef.name
```

Paths must resolve directly to a Secret-name string.

## Multiple API versions

```yaml
apiVersions:
  - external-secrets.io/v1
  - external-secrets.io/v1beta1
```

`kubectl-peek` uses Kubernetes discovery and selects the first available version. If none is installed, the rule is skipped.

## Loading rules

Explicitly:

```bash
kubectl peek secret --rules ./rules.yaml
```

Persistently:

```bash
export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"
```

Suggested setup:

```bash
mkdir -p "$HOME/.config/kubectl-peek"
cp examples/rules-all.yaml "$HOME/.config/kubectl-peek/rules.yaml"
echo 'export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"' >> "$HOME/.zshrc"
source "$HOME/.zshrc"
```

Precedence:

1. `--rules`
2. `KUBECTL_PEEK_RULE_FILE`
3. built-in discovery only

The repository includes examples under `examples/`, including External Secrets, cert-manager, Crossplane, common resources and a combined rules file.
