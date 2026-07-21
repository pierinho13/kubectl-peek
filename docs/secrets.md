# Secret inspection

[← Back to README](../README.md)

Select and inspect Secrets in the current namespace:

```bash
kubectl peek secret
```

![kubectl-peek Secret inspection](https://github.com/user-attachments/assets/02ec34c5-f82a-443c-a3f1-a03a2c9d5489)

The selector supports pagination, filtering and keyboard navigation. Press `Enter` to inspect the highlighted Secret.

## Common commands

```bash
kubectl peek secret database
kubectl peek secret -n staging
kubectl peek secret database -n staging
kubectl peek secret --context development-cluster
kubectl peek secret --kubeconfig ~/.kube/secondary-config
kubectl peek secret database \
  --context development-cluster \
  --namespace staging \
  --kubeconfig ~/.kube/secondary-config
```

Matching by name is case-insensitive.

Aliases:

```bash
kubectl peek secrets
kubectl peek sec
```

## Built-in relationship discovery

Without custom configuration, `kubectl-peek` searches supported resources in the selected namespace, including:

- Pods
- Deployments
- StatefulSets
- DaemonSets
- Jobs
- CronJobs
- ServiceAccounts
- Ingresses
- Gateway API Gateways, when installed

Built-in finders detect references such as:

- `env[].valueFrom.secretKeyRef`
- `envFrom[].secretRef`
- Secret-backed and projected volumes
- `imagePullSecrets`
- init-container and ephemeral-container references
- ServiceAccount pull Secrets
- TLS Secret references

Example:

```text
Secret: database-credentials
Namespace: default
Type: Opaque
Used by:
  Deployment/backend
    uses: environment variable
  CronJob/backup
    uses: environment variable
```

When no supported resource references the Secret:

```text
Used by:
  none
```

This is useful dependency information, not a guarantee that deleting a Secret is safe. A Secret can be consumed externally or through an unsupported path.

## Custom resources

Load a declarative rule file:

```bash
kubectl peek secret --rules ./rules.yaml
kubectl peek secret --rules ./examples/rules-all.yaml
```

See [Custom discovery rules](rules.md).
