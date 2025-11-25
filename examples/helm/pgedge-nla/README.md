# pgEdge Natural Language Agent - Helm Chart

This Helm chart deploys the pgEdge Natural Language Agent on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PostgreSQL database (external or in-cluster)
- Optional: cert-manager for TLS certificates

## Installation

```bash
# Install from local chart
helm install pgedge-nla ./pgedge-nla \
  --namespace pgedge \
  --create-namespace \
  --set secrets.postgresPassword="your-secure-password"

# Install with production values
helm install pgedge-nla ./pgedge-nla \
  --namespace pgedge \
  --create-namespace \
  -f values-production.yaml

# Upgrade
helm upgrade pgedge-nla ./pgedge-nla \
  --namespace pgedge \
  -f values-production.yaml

# Uninstall
helm uninstall pgedge-nla --namespace pgedge
```

## Configuration

See [values.yaml](values.yaml) for all configuration options.

### Key Configuration Options

- `server.replicaCount`: Number of server replicas (default: 2)
- `server.resources`: CPU and memory limits
- `server.autoscaling.enabled`: Enable horizontal pod autoscaling
- `ingress.enabled`: Enable ingress for external access
- `secrets.postgresPassword`: PostgreSQL password (required)

## Production Deployment

For production deployments, use the provided `values-production.yaml` which
includes:

- Higher replica counts
- Resource limits and autoscaling
- Pod anti-affinity for high availability
- Ingress with TLS

See [values-production.yaml](values-production.yaml) for details.
