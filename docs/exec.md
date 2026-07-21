# Interactive Pod exec

[← Back to README](../README.md)

Select a Pod interactively and open a shell inside one of its containers:

```bash
kubectl peek exec
```

![kubectl-peek exec](https://github.com/user-attachments/assets/1a46465c-e129-4a63-9200-0658fed6de94)

The selector displays the Pod name, readiness, container count and phase. Completed and failed Pods are excluded.

```text
Select Pod
    ↓
Select container when more than one exists
    ↓
Open /bin/bash or fall back to /bin/sh
```

## Common commands

```bash
kubectl peek exec api
kubectl peek exec -n staging
kubectl peek exec api \
  --namespace staging \
  --container application \
  --shell /bin/sh
```

Overrides are available for namespace, context, kubeconfig, container and shell.

Alias:

```bash
kubectl peek x
```

Exit the container shell with:

```bash
exit
```
