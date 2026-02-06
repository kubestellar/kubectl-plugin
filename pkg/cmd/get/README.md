# Get Command Package Structure

This package contains the decoupled implementation of the `kubectl multi get` command, organized by Kubernetes API groups.

## Directory Structure

```
pkg/cmd/get/
├── types.go           # Shared types and interfaces
├── common.go          # Common utility functions
├── core/              # Core API group resources (pods, services, configmaps, secrets, etc.)
├── apps/              # Apps API group resources (deployments, replicasets, statefulsets, daemonsets)
├── batch/             # Batch API group resources (jobs, cronjobs)
├── networking/        # Networking API group resources (ingresses, networkpolicies)
├── rbac/              # RBAC API group resources (roles, rolebindings)
└── storage/           # Storage API group resources (storageclasses)
```

## Guidelines for Adding New Resource Handlers

1. Each resource type should have its own file named after the resource (e.g., `pods.go`, `deployments.go`)
2. Place the file in the appropriate directory based on its Kubernetes API group
3. Use the shared types and utilities from `types.go` and `common.go`
4. Follow the existing function signature pattern: `handle<Resource>Get(...)`
5. Ensure all functions are in the `cmd` package to maintain compatibility with the main `get.go` file

## Migration Progress

This is part of the effort to decouple the large `get.go` file (2,514 lines) into smaller, more maintainable files.

**Status:** Infrastructure setup complete ✅

**Next steps:** Move resource handlers to appropriate subdirectories in subsequent PRs.
