# BindingPolicy Management Guide

BindingPolicies are the core mechanism in KubeStellar for defining how workloads are distributed across multiple clusters. This guide covers comprehensive management of BindingPolicies using the KubeStellar CLI tools.

## Overview

A BindingPolicy defines:
- **Which clusters** should receive workloads (cluster selectors)
- **Which resources** should be distributed (downsync specifications)
- **How conflicts** should be resolved
- **What customizations** should be applied per cluster

## BindingPolicy Structure

### Basic Structure

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: example-policy
  namespace: default
spec:
  clusterSelectors:
  - matchLabels:
      type: "wec"
  downsync:
  - apiVersion: apps/v1
    kind: Deployment
    namespaced: true
  - apiVersion: v1
    kind: Service
    namespaced: true
```

### Key Components

1. **clusterSelectors**: Define which clusters receive workloads
2. **downsync**: Specify which resource types to distribute
3. **customizations**: Apply cluster-specific modifications (optional)
4. **constraints**: Define placement constraints (optional)

## Using the CLI Tools

### Creating BindingPolicies

#### Basic Creation

Using the standalone CLI:
```bash
# Interactive mode
kubestellar
kubestellar> bindingpolicy create my-app-policy

# Direct command
kubestellar bindingpolicy create my-app-policy \
  --clusters cluster1,cluster2 \
  --namespace production
```

Using the kubectl plugin:
```bash
# Not directly supported - use YAML files with apply
kubectl multi apply -f my-bindingpolicy.yaml
```

#### Advanced Creation with Options

```bash
# Create policy for all clusters
kubestellar bindingpolicy create global-policy --all-clusters

# Create policy with specific cluster labels
kubestellar bindingpolicy create labeled-policy \
  --match-labels "type=wec,env=production"

# Create namespaced policy
kubestellar bindingpolicy create app-policy \
  --namespace my-app \
  --clusters cluster1,cluster2 \
  --namespaced
```

### Listing BindingPolicies

#### View All Policies

```bash
# Using standalone CLI
kubestellar bindingpolicy list

# Using kubectl plugin (viewing the CRDs)
kubectl multi get bindingpolicies -A
```

Example output:
```
NAMESPACE    NAME           CLUSTERS
default      app-policy     all
production   prod-policy    2 selector(s)
staging      test-policy    cluster1,cluster2
```

#### Detailed View

```bash
# Get detailed information about a specific policy
kubestellar bindingpolicy describe my-policy

# View YAML representation
kubectl multi get bindingpolicy my-policy -o yaml
```

### Updating BindingPolicies

#### Adding Clusters

```bash
# Add clusters to an existing policy
kubestellar bindingpolicy add-clusters my-policy cluster3 cluster4

# Interactive mode
kubestellar> bindingpolicy add-clusters my-policy cluster3 cluster4
```

#### Removing Clusters

```bash
# Remove clusters from a policy
kubestellar bindingpolicy remove-clusters my-policy cluster1

# Interactive mode
kubestellar> bindingpolicy remove-clusters my-policy cluster1
```

#### Updating Labels and Selectors

```bash
# Update cluster selector labels
kubectl multi patch bindingpolicy my-policy --type='merge' -p='{
  "spec": {
    "clusterSelectors": [
      {
        "matchLabels": {
          "type": "wec",
          "env": "production",
          "zone": "us-east-1"
        }
      }
    ]
  }
}'
```

### Deleting BindingPolicies

```bash
# Using standalone CLI
kubestellar bindingpolicy delete my-policy

# Using kubectl plugin
kubectl multi delete bindingpolicy my-policy

# Interactive mode
kubestellar> bindingpolicy delete my-policy
```

## Common Patterns

### 1. Application-Specific Policies

Create policies for specific applications:

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-app-policy
  namespace: web-apps
spec:
  clusterSelectors:
  - matchLabels:
      app-tier: "web"
      env: "production"
  downsync:
  - apiVersion: apps/v1
    kind: Deployment
    namespaced: true
    name: "nginx-*"
  - apiVersion: v1
    kind: Service
    namespaced: true
    name: "nginx-*"
  - apiVersion: v1
    kind: ConfigMap
    namespaced: true
    name: "nginx-*"
```

### 2. Environment-Based Policies

Separate policies for different environments:

```bash
# Production policy
kubestellar bindingpolicy create production-policy \
  --match-labels "env=production" \
  --namespace production

# Staging policy  
kubestellar bindingpolicy create staging-policy \
  --match-labels "env=staging" \
  --namespace staging

# Development policy
kubestellar bindingpolicy create development-policy \
  --match-labels "env=development" \
  --namespace development
```

### 3. Resource-Type Specific Policies

Create policies for specific resource types:

```yaml
# Policy for stateless applications
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: stateless-apps-policy
spec:
  clusterSelectors:
  - matchLabels:
      type: "wec"
  downsync:
  - apiVersion: apps/v1
    kind: Deployment
    namespaced: true
  - apiVersion: v1
    kind: Service
    namespaced: true

---
# Policy for stateful applications
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: stateful-apps-policy
spec:
  clusterSelectors:
  - matchLabels:
      type: "wec"
      storage: "available"
  downsync:
  - apiVersion: apps/v1
    kind: StatefulSet
    namespaced: true
  - apiVersion: v1
    kind: PersistentVolumeClaim
    namespaced: true
```

### 4. Global Infrastructure Policies

Policies for cluster-wide resources:

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: global-infrastructure-policy
spec:
  clusterSelectors:
  - matchLabels:
      type: "wec"
  downsync:
  - apiVersion: v1
    kind: Namespace
    namespaced: false
  - apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    namespaced: false
  - apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    namespaced: false
```

## Advanced Features

### 1. Cluster Customizations

Apply different configurations to different clusters:

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: customized-app-policy
spec:
  clusterSelectors:
  - matchLabels:
      type: "wec"
  downsync:
  - apiVersion: apps/v1
    kind: Deployment
    namespaced: true
  customizations:
  - clusters:
    - "us-east-cluster"
    patches:
    - type: "strategic"
      patch: |
        spec:
          replicas: 5
          template:
            spec:
              nodeSelector:
                zone: "us-east-1"
  - clusters:
    - "us-west-cluster"
    patches:
    - type: "strategic"
      patch: |
        spec:
          replicas: 3
          template:
            spec:
              nodeSelector:
                zone: "us-west-1"
```

### 2. Conditional Policies

Policies with complex conditions:

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: conditional-policy
spec:
  clusterSelectors:
  - matchLabels:
      type: "wec"
      cpu-arch: "amd64"
  - matchExpressions:
    - key: "kubernetes-version"
      operator: "In"
      values: ["1.28", "1.29", "1.30"]
    - key: "node-count"
      operator: "Gt"
      values: ["2"]
  downsync:
  - apiVersion: apps/v1
    kind: Deployment
    namespaced: true
```

## Monitoring and Troubleshooting

### 1. Check Policy Status

```bash
# View policy status across clusters
kubectl multi describe bindingpolicy my-policy

# Check if resources are being distributed
kubectl multi get deployments -l kubestellar.io/binding-policy=my-policy
```

### 2. Verify Cluster Selection

```bash
# List clusters that match policy selectors
kubestellar clusters list --match-labels "type=wec,env=production"

# Check cluster labels
kubestellar clusters info my-cluster
```

### 3. Debug Distribution Issues

```bash
# Check for binding policy events
kubectl multi get events | grep BindingPolicy

# Verify resource creation in target clusters
kubectl multi get deployments -A | grep my-app

# Check for conflicts or errors
kubectl multi describe bindingpolicy my-policy | grep -i error
```

### 4. Common Issues and Solutions

#### Issue: Resources not being distributed

**Cause**: Incorrect cluster selectors

**Solution**:
```bash
# Check cluster labels
kubestellar clusters list

# Update policy selectors
kubestellar bindingpolicy update my-policy \
  --match-labels "correct-label=correct-value"
```

#### Issue: Policy conflicts

**Cause**: Multiple policies selecting the same resources

**Solution**:
```bash
# List all policies
kubestellar bindingpolicy list

# Check for overlapping selectors
kubestellar bindingpolicy analyze-conflicts
```

#### Issue: Permissions errors

**Cause**: Insufficient RBAC permissions

**Solution**:
```bash
# Check permissions in target clusters
kubectl multi auth can-i create deployments --all-namespaces

# Verify service account configuration
kubectl multi get serviceaccount kubestellar-agent -A
```

## Best Practices

### 1. Naming Conventions

Use descriptive names that indicate purpose and scope:

```bash
# Good examples
kubestellar bindingpolicy create web-tier-production
kubestellar bindingpolicy create monitoring-infrastructure  
kubestellar bindingpolicy create db-tier-stateful

# Avoid generic names
kubestellar bindingpolicy create policy1  # Bad
kubestellar bindingpolicy create test     # Bad
```

### 2. Granular Policies

Create specific policies rather than overly broad ones:

```bash
# Preferred: Specific policies
kubestellar bindingpolicy create frontend-apps --match-labels "tier=frontend"
kubestellar bindingpolicy create backend-apis --match-labels "tier=backend"

# Avoid: Overly broad policies
kubestellar bindingpolicy create everything --all-clusters  # Too broad
```

### 3. Environment Separation

Maintain separate policies for different environments:

```bash
# Environment-specific policies
kubestellar bindingpolicy create prod-web-apps \
  --match-labels "env=production,tier=web" \
  --namespace production

kubestellar bindingpolicy create staging-web-apps \
  --match-labels "env=staging,tier=web" \
  --namespace staging
```

### 4. Resource Optimization

Be specific about which resources to distribute:

```yaml
# Preferred: Specific resource selection
downsync:
- apiVersion: apps/v1
  kind: Deployment
  namespaced: true
  name: "web-*"
- apiVersion: v1
  kind: Service
  namespaced: true
  name: "web-*"

# Avoid: Overly broad selections  
downsync:
- apiVersion: "*"
  kind: "*"
  namespaced: true  # Too broad
```

### 5. Documentation and Labeling

Document policies with labels and annotations:

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: web-app-policy
  labels:
    app: "web-application"
    team: "frontend"
    environment: "production"
  annotations:
    description: "Distributes web application components to production clusters"
    owner: "frontend-team@company.com"
    created-by: "deployment-automation"
spec:
  # ... policy specification
```

## Integration Examples

### 1. CI/CD Pipeline Integration

```bash
#!/bin/bash
# deploy.sh - Deployment script

# Create binding policy for new application
kubestellar bindingpolicy create ${APP_NAME}-policy \
  --match-labels "env=${ENVIRONMENT},tier=${APP_TIER}" \
  --namespace ${APP_NAMESPACE}

# Apply application manifests
kubectl multi apply -f ./k8s/

# Verify distribution
kubectl multi get deployments -l app=${APP_NAME} --all-namespaces
```

### 2. GitOps Integration

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- bindingpolicy.yaml
- deployment.yaml
- service.yaml

# Apply with ArgoCD or similar GitOps tool
# The binding policy ensures distribution to appropriate clusters
```

### 3. Helm Chart Integration

```yaml
# templates/bindingpolicy.yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: {{ include "myapp.fullname" . }}-policy
  namespace: {{ .Values.namespace }}
spec:
  clusterSelectors:
  {{- range .Values.clusterSelectors }}
  - matchLabels:
    {{- toYaml .matchLabels | nindent 6 }}
  {{- end }}
  downsync:
  - apiVersion: apps/v1
    kind: Deployment
    namespaced: true
    name: {{ include "myapp.fullname" . }}
```

## See Also

- [Cluster Management](cluster_management.md) - Managing cluster registration and labels
- [Helm Integration](helm_integration.md) - Using Helm with BindingPolicies
- [Architecture Guide](architecture_guide.md) - Understanding KubeStellar architecture
- [Usage Guide](usage_guide.md) - General usage patterns and examples