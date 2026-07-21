# Permissions and security

[← Back to README](../README.md)

`kubectl-peek` uses the active Kubernetes identity. It never bypasses RBAC.

## Required permissions

### Secret inspection

```yaml
apiGroups: [""]
resources: ["secrets"]
verbs: ["get", "list"]
```

Relationship discovery also requires `list` access to the resource types being searched, such as Pods, ServiceAccounts, Deployments, StatefulSets, DaemonSets, Jobs, CronJobs, Ingresses, Gateways and any resources configured through custom rules.

### Event inspection

```yaml
apiGroups: [""]
resources: ["events"]
verbs: ["list"]
```

### Pod exec

```yaml
apiGroups: [""]
resources: ["pods"]
verbs: ["get", "list"]
---
apiGroups: [""]
resources: ["pods/exec"]
verbs: ["create"]
```

## Verify access

```bash
kubectl auth can-i list secrets -n default
kubectl auth can-i get secrets -n default
kubectl auth can-i list deployments.apps -n default
kubectl auth can-i list events -n default
kubectl auth can-i list pods -n default
kubectl auth can-i create pods/exec -n default
```

Custom rules require `list` access to each configured resource.

## Secret-value warning

`kubectl-peek` prints decoded Secret values directly to the terminal. Values may remain visible in:

- terminal scrollback;
- shell session recordings;
- shared terminal sessions;
- screen-sharing sessions;
- screenshots;
- captured command output;
- CI logs.

Use the feature only in trusted environments. Never paste real Secret values into public issues or pull requests.

## Isolated-shell temporary kubeconfigs

Temporary kubeconfigs:

- are created with `0600` permissions;
- are available only to the child shell through `KUBECONFIG`;
- preserve the effective authentication configuration;
- do not change the original kubeconfig;
- are removed after normal shell exit.

Because they contain Kubernetes credentials and embedded certificate/key data, treat the temporary shell environment as sensitive.
