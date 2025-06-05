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
	colFmt       = "%-6s | %-20s | %-15s | %-28s\n" // format for tabular output
	separatorLen = 78                               // length of the table separator line
)

func main() {
	// Command-line flags
	remoteCtx := flag.String("remote-context", "its1", "remote hosting context (for ManagedCluster)")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (defaults to $HOME/.kube/config)")
	flag.Parse()

	// ---------- Local Cluster ----------
	currCtx, localClient := buildClient(*kubeconfig, "")
	printHeader()
	printRow(currCtx, currCtx, "CLUSTER", "-")
	listLocalNodes(localClient)

	// ---------- Remote Cluster ----------
	if *remoteCtx != "" {
		remoteClient := buildDynamicClient(*kubeconfig, *remoteCtx)
		listManagedClusters(remoteClient, *remoteCtx)
	}
}

// buildClient returns the current context name and a kubernetes Clientset
func buildClient(kcfg, overrideCtx string) (string, *kubernetes.Clientset) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfg != "" {
		loadingRules.ExplicitPath = kcfg
	}
	cfgOverrides := &clientcmd.ConfigOverrides{}
	if overrideCtx != "" {
		cfgOverrides.CurrentContext = overrideCtx
	}
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, cfgOverrides)
	rawCfg, err := cc.RawConfig()
	exitIf(err)

	ctx := rawCfg.CurrentContext
	restCfg, err := cc.ClientConfig()
	exitIf(err)

	clientset, err := kubernetes.NewForConfig(restCfg)
	exitIf(err)
	return ctx, clientset
}

// buildDynamicClient returns a dynamic.Interface client for custom resources
func buildDynamicClient(kcfg, overrideCtx string) dynamic.Interface {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfg != "" {
		loadingRules.ExplicitPath = kcfg
	}
	cfgOverrides := &clientcmd.ConfigOverrides{CurrentContext: overrideCtx}
	restCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, cfgOverrides).ClientConfig()
	exitIf(err)

	dyn, err := dynamic.NewForConfig(restCfg)
	exitIf(err)
	return dyn
}

// listLocalNodes prints nodes in the current local cluster
func listLocalNodes(cs *kubernetes.Clientset) {
	ctx := context.TODO()
	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	exitIf(err)

	for _, n := range nodes.Items {
		age := humanAge(n.CreationTimestamp.Time)
		printRow("", "  "+n.Name, "NODE", age)
	}
}

// listManagedClusters lists remote managed clusters from the specified context
func listManagedClusters(dyn dynamic.Interface, remoteCtx string) {
	ctx := context.TODO()
	gvr := schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}
	mcs, err := dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: list managedclusters failed in context %s: %v\n", remoteCtx, err)
		return
	}
	for _, item := range mcs.Items {
		age := humanAge(item.GetCreationTimestamp().Time)
		printRow(remoteCtx, item.GetName(), "REMOTE-CLUSTER", age)
	}
}

// printHeader prints the table header
func printHeader() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, colFmt, "CTX", "RESOURCE", "TYPE", "AGE")
	fmt.Fprintln(w, strings.Repeat("-", separatorLen))
	w.Flush()
}

// printRow prints a formatted row
func printRow(a, b, c, d string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, colFmt, a, b, c, d)
	w.Flush()
}

// humanAge returns a human-readable age string from the given time
func humanAge(t time.Time) string {
	d := time.Since(t).Round(time.Second)
	return d.String()
}

// exitIf prints the error and exits if the error is non-nil
func exitIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
