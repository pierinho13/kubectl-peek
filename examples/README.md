# kubectl-peek rule examples

This directory contains example rule files for extending Secret relationship discovery.

## Files

- `rules-external-secrets.yaml`: External Secrets Operator examples.
- `rules-cert-manager.yaml`: cert-manager `Certificate` support.
- `rules-crossplane.yaml`: Crossplane-style connection Secret examples.
- `rules-common.yaml`: mixed examples for common and internal CRDs.
- `rules-all.yaml`: combined starter file.

## Use a file directly

```bash
kubectl-peek --rules ./examples/rules-all.yaml
```

## Configure a persistent default

```bash
mkdir -p "$HOME/.config/kubectl-peek"
cp ./examples/rules-all.yaml "$HOME/.config/kubectl-peek/rules.yaml"

export KUBECTL_PEEK_RULE_FILE="$HOME/.config/kubectl-peek/rules.yaml"
```

Add the export to `~/.zshrc`, `~/.bashrc`, or the corresponding shell configuration file.

## Rule structure

```yaml
rules:
  - apiVersions:
      - example.io/v1
    kind: Example
    resource: examples
    references:
      - path: spec.credentials.secretRef.name
        relation: uses
        description: application credentials
```

## Important considerations

- `resource` must be the plural API resource name.
- `path` must resolve directly to a Secret name.
- Use `*` to traverse arrays.
- List multiple API versions in preferred order.
- Remove example rules that do not apply to your cluster.
- Verify all paths against real manifests or installed CRDs.
- Custom discovery requires permission to list the configured resource.

Check installed resources with:

```bash
kubectl api-resources
```

Inspect a CRD schema with:

```bash
kubectl get crd <name> -o yaml
```

Test permissions with:

```bash
kubectl auth can-i list <resource>.<api-group> -n <namespace>
```
