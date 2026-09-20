# tokeninfra Kubernetes deployment

This directory records the manifests currently used by the three-node CCE deployment.

## Prerequisites

- Nodes are named `172.20.0.5`, `172.20.0.6`, and `172.20.0.7`.
- The CCE load balancer `lb-3k3btl3b` already exists.
- The new-api image is available as `172.20.0.7:5000/new-api:tokeninfra-brand-20260828`.
- Replace every value in `secrets.example.yaml` before applying it. Do not commit real credentials.

## Apply

Create namespaces and the registry first:

```bash
kubectl apply -f deploy/tokeninfra/namespaces.yaml
kubectl apply -f deploy/tokeninfra/registry.yaml
kubectl -n registry-system rollout status deployment/local-registry
kubectl -n registry-system rollout status daemonset/local-registry-config
```

Create the two runtime Secrets from a private copy:

```bash
cp deploy/tokeninfra/secrets.example.yaml /tmp/tokeninfra-secrets.yaml
# Edit /tmp/tokeninfra-secrets.yaml without copying credentials into the repository.
kubectl apply -f /tmp/tokeninfra-secrets.yaml
```

Deploy Redis and new-api:

```bash
kubectl apply -f deploy/tokeninfra/redis.yaml
kubectl -n redis rollout status deployment/redis-0
kubectl apply -f deploy/tokeninfra/new-api.yaml
kubectl -n new-api rollout status deployment/new-api
```

Redis is cache-only: RDB and AOF are disabled, and the Service is `ClusterIP` only.
