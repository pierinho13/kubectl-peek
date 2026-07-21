# Kubernetes Event investigation

[← Back to README](../README.md)

Browse Events in the current namespace:

```bash
kubectl peek events
```

![kubectl-peek Event inspection](https://github.com/user-attachments/assets/ca59b59a-ab7c-4140-8e50-10b3f38b0efd)

Events are ordered by their latest occurrence. Repeated records are grouped by default. The list shows the latest occurrence, type, reason, involved object and total occurrences.

Press `Enter` to inspect:

- exact occurrence count;
- first-seen and last-seen timestamps;
- source;
- namespace;
- involved object;
- full message.

## Filters

```bash
kubectl peek events --warnings
kubectl peek events --non-normal
kubectl peek events --no-group
kubectl peek events backoff
kubectl peek events -n staging
kubectl peek events -A
```

## Browse by resource

```bash
kubectl peek events --warnings --browse
kubectl peek events --warnings --browse-by-kind
```

Navigation:

```text
Kind → Resource → Events → Detail
```

The first screen aggregates resources, Events and occurrences by Kubernetes kind. The next screen shows resources of the selected kind, followed by matching Events and their details.

Browse mode combines with other filters:

```bash
kubectl peek events --non-normal --browse
kubectl peek events backoff --browse
kubectl peek events -A --warnings --browse
kubectl peek events --no-group --browse
```

Aliases:

```bash
kubectl peek event
kubectl peek ev
```
