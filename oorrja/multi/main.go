// This CLI plugin allows you to list nodes from the current Kubernetes cluster
// and from remote managed clusters registered via OCM (Open Cluster Management).
// Usage: kubectl multi get nodes [--remote-context <name>] [--kubeconfig <path>]

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	tableFmt     = "%-8s | %-30s | %-8s | %-12s | %-10s | %-10s\n"
	separatorLen = 90
)

// main is the entry point. It sets up clients for local and remote clusters,
// prints the local node table, and queries remote managed clusters if provided.
func main() {
	remoteCtx := flag.String("remote-context", "its1", "remote hosting context (for ManagedCluster)")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig")
	flag.Parse()

	currCtx, localClient := buildClient(*kubeconfig, "")
	printTableHeader()
	printTableSeparator()

	printTableRow(currCtx, currCtx, "Ready", "CLUSTER", "-", "-")
	listNodes("", localClient)

	if *remoteCtx != "" {
		remoteClient := buildDynamicClient(*kubeconfig, *remoteCtx)
		listManagedClusters(remoteClient, *remoteCtx, *kubeconfig, currCtx)
	}
}

// buildClient builds a Kubernetes clientset using a kubeconfig path and optional context override.
// returns the context name and clientset.
func buildClient(kcfg, ctxOverride string) (string, *kubernetes.Clientset) {
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfg != "" {
		loading.ExplicitPath = kcfg
	}
	overrides := &clientcmd.ConfigOverrides{}
	if ctxOverride != "" {
		overrides.CurrentContext = ctxOverride
	}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	raw, err := cfg.RawConfig()
	exitIf(err)

	restCfg, err := cfg.ClientConfig()
	exitIf(err)

	clientset, err := kubernetes.NewForConfig(restCfg)
	exitIf(err)

	return raw.CurrentContext, clientset
}

// buildDynamicClient builds a dynamic Kubernetes client for interacting with custom resources,
// like ManagedClusters, using a specific context.
func buildDynamicClient(kcfg, ctxOverride string) dynamic.Interface {
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfg != "" {
		loading.ExplicitPath = kcfg
	}
	overrides := &clientcmd.ConfigOverrides{CurrentContext: ctxOverride}
	restCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides).ClientConfig()
	exitIf(err)

	dyn, err := dynamic.NewForConfig(restCfg)
	exitIf(err)
	return dyn
}

// listNodes retrieves and prints details of nodes (name, status, role, age, version)
// in a given cluster context using the provided clientset.
func listNodes(ctxName string, cs *kubernetes.Clientset) {
	nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	exitIf(err)

	for _, n := range nodes.Items {
		status := "Unknown"
		for _, cond := range n.Status.Conditions {
			if cond.Type == "Ready" {
				status = string(cond.Status)
				break
			}
		}

		role := "<none>"
		for k := range n.Labels {
			if strings.HasPrefix(k, "node-role.kubernetes.io/") {
				role = strings.TrimPrefix(k, "node-role.kubernetes.io/")
				break
			}
		}

		version := n.Status.NodeInfo.KubeletVersion
		age := humanAge(n.CreationTimestamp.Time)
		printTableRow("", "└─ "+n.Name, status, role, age, version)
	}
}

// listManagedClusters fetches managed clusters from a remote context and prints their metadata.
// It then queries and prints node data for each cluster (excluding the local context).
func listManagedClusters(dyn dynamic.Interface, remoteCtx, kubeconfig, localCtx string) {
	gvr := schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}
	clusters, err := dyn.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not list managedclusters from %s: %v\n", remoteCtx, err)
		return
	}
	for _, mc := range clusters.Items {
		name := mc.GetName()
		if name == localCtx {
			continue
		}
		age := humanAge(mc.GetCreationTimestamp().Time)
		printTableRow(remoteCtx, name, "Ready", "CLUSTER", age, "-")
		_, client := buildClient(kubeconfig, name)
		listNodes(name, client)
	}
}

// printTableHeader prints the header row of the output table.
func printTableHeader() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, tableFmt, "CTX", "NAME", "STATUS", "ROLES", "AGE", "VERSION")
	w.Flush()
}

func printTableSeparator() {
	fmt.Println(strings.Repeat("-", separatorLen))
}

// printTableRow prints a formatted row of cluster/node data in the table.
func printTableRow(ctx, name, status, roles, age, version string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, tableFmt, ctx, name, status, roles, age, version)
	w.Flush()
}

// humanAge returns a human-readable string showing the age of a resource from its creation time.
func humanAge(t time.Time) string {
	return time.Since(t).Round(time.Second).String()
}

// exitIf prints an error and exits the program if the given error is non-nil.
func exitIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

