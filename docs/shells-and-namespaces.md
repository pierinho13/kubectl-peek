# Shells and namespaces

[← Back to README](../README.md)

## Open an isolated Kubernetes shell

```bash
kubectl peek shell
```

![kubectl-peek shell workflows](https://github.com/user-attachments/assets/94bd4178-0ecf-4750-bc31-1e46e155200b)

Interactive flow:

```text
Select context
      ↓
List namespaces through that context
      ↓
Select namespace
      ↓
Create a temporary kubeconfig
      ↓
Open an isolated child shell
```

A context is selected rather than a cluster because a Kubernetes context combines the cluster and the credentials used to access it.

Skip either selector by passing values directly:

```bash
kubectl peek shell --context operations
kubectl peek shell --namespace traefik
kubectl peek shell -n traefik
kubectl peek shell --context operations --namespace traefik
kubectl peek shell --kubeconfig ~/.kube/secondary-config
```

A supplied namespace is validated through the selected context before the shell starts.

## Isolation behavior

The original kubeconfig and current context are not modified. The selected scope applies only to the child shell through a temporary `KUBECONFIG`.

```text
┌─ kubectl-peek isolated shell
│ Context    operations
│ Namespace  traefik
└─ Type `exit` to return to the previous shell
```

The child prompt keeps the scope visible:

```text
[k8s:operations ns:traefik] user@host %
```

Commands inside the shell use the isolated configuration:

```bash
kubectl get pods
kubectl get services
helm list
flux get all
```

The temporary kubeconfig:

- contains the complete effective configuration;
- preserves clusters, users, contexts, authentication and exec plugins;
- embeds certificate and key data referenced by file paths;
- changes only the selected context namespace;
- is created with `0600` permissions;
- is exposed only to the child shell;
- does not modify the original kubeconfig.

Run `exit` to return and remove the temporary directory.

Nested isolated shells are blocked before opening any selector:

```text
an isolated kubectl-peek shell is already active
(context "operations", namespace "traefik");
run exit before opening another one
```

## Work across clusters in parallel

Open separate terminal windows:

```bash
kubectl peek shell --context staging -n payments
kubectl peek shell --context production -n payments
```

Each terminal keeps an independent Kubernetes scope.

## Select and persist a namespace

```bash
kubectl peek namespace
kubectl peek ns
```

![kubectl-peek namespace workflow](https://github.com/user-attachments/assets/9eb0f303-cc02-4cc2-b434-ead0ceb10c35)

The selector supports keyboard navigation, pagination and filtering. After selection, the namespace stored in the chosen kubeconfig context is updated.

```bash
kubectl peek namespace traefik
kubectl peek namespace --context staging
kubectl peek namespace --kubeconfig ~/.kube/secondary-config
```

## Namespace-first isolated shell

```bash
kubectl peek namespace --shell
kubectl peek namespace traefik --shell
kubectl peek namespace --context staging --shell
```

![kubectl-peek namespace shell](https://github.com/user-attachments/assets/fae7e6dd-1a36-491b-a4fe-0cbe5eaed4ee)

This combines namespace selection with an isolated child shell without changing the original kubeconfig.
