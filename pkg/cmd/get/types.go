package cmd

import (
	"text/tabwriter"

	"kubectl-multi/pkg/cluster"
)

// ResourceHandler defines the interface for handling get operations for different resource types
type ResourceHandler interface {
	Handle(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error
}

// ClusterScopedResourceHandler defines the interface for cluster-scoped resources (nodes, namespaces, etc.)
type ClusterScopedResourceHandler interface {
	Handle(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error
}

// GetOptions contains common options for get operations
type GetOptions struct {
	OutputFormat  string
	Selector      string
	ShowLabels    bool
	Watch         bool
	WatchOnly     bool
	Namespace     string
	AllNamespaces bool
	Kubeconfig    string
	RemoteCtx     string
}
