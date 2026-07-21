# Development and troubleshooting

[← Back to README](../README.md)

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build -o kubectl-peek .
```

Run locally:

```bash
./kubectl-peek
./kubectl-peek secret
./kubectl-peek events --warnings --browse
./kubectl-peek exec
./kubectl-peek shell
```

Test native plugin usage by placing the binary in your `PATH`:

```bash
kubectl peek
kubectl peek secret
kubectl peek events --warnings
kubectl peek exec
kubectl peek namespace
kubectl peek shell
```

## Troubleshooting custom rules

Check the configured path:

```bash
echo "$KUBECTL_PEEK_RULE_FILE"
ls -l "$KUBECTL_PEEK_RULE_FILE"
```

Pass it explicitly:

```bash
kubectl peek secret --rules "$HOME/.config/kubectl-peek/rules.yaml"
```

When a CRD is not detected, verify:

1. `apiVersions` matches an installed version.
2. `resource` is the plural API resource name.
3. `path` resolves directly to the Secret name.
4. The object exists in the selected namespace.
5. The active identity can list it.
6. The YAML is valid.

Useful commands:

```bash
kubectl api-resources
kubectl get <resource> -n <namespace>
kubectl auth can-i list <resource> -n <namespace>
```

## No Secret relationships are displayed

Possible causes:

- no supported resource references it;
- the resource is in another namespace;
- the field path is unsupported;
- a custom rule is missing;
- RBAC prevents listing the resource;
- the Secret is consumed outside Kubernetes.

Discovery is helpful context, not a deletion-safety guarantee.

## No Events are displayed

```bash
kubectl get events -n default
kubectl auth can-i list events -n default
```

Check whether filters are excluding all results.

## Pod exec fails

```bash
kubectl get pods -n default
kubectl auth can-i create pods/exec -n default
kubectl exec -n default <pod> -- /bin/sh
```

Verify that the Pod is running, the container exists and the requested shell is installed.

## Verify an isolated shell

Inside the shell:

```bash
echo "$KUBECONFIG"
kubectl config current-context
kubectl config view --minify -o jsonpath='{..namespace}{"\n"}'
```

The prompt should also display the selected context and namespace. Run `exit` to return and clean up temporary files.
