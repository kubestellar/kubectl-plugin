package cmd

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/duration"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"
)

// Custom help function for get command
func getHelpFunc(cmd *cobra.Command, args []string) {
	// Get original kubectl help using the new implementation
	cmdInfo, err := util.GetKubectlCommandInfo("get")
	if err != nil {
		// Fallback to default help if kubectl help is not available
		cmd.Help()
		return
	}

	// Multi-cluster plugin information
	multiClusterInfo := `Get resources from all managed clusters and display them in a unified view.
Supports all resource types that kubectl get supports.

The output includes cluster context information to help identify which
cluster each resource belongs to.`

	// Multi-cluster examples
	multiClusterExamples := `# List all pods in all managed clusters
kubectl multi get pods

# List all nodes in all managed clusters
kubectl multi get nodes

# List deployments in specific namespace across all clusters
kubectl multi get deployments -n production

# List all resources in all namespaces across all clusters
kubectl multi get all -A

# Get pods with labels
kubectl multi get pods -l app=nginx

# Get specific pod across all clusters
kubectl multi get pod nginx-pod

# Get services with wide output
kubectl multi get services -o wide
 
#get all job
kubectl multi get jobs
`

	// Multi-cluster usage
	multiClusterUsage := `kubectl multi get [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...] [flags]`

	// Format combined help using the new CommandInfo structure
	combinedHelp := util.FormatMultiClusterHelp(cmdInfo, multiClusterInfo, multiClusterExamples, multiClusterUsage)
	fmt.Fprintln(cmd.OutOrStdout(), combinedHelp)
}

func newGetCommand() *cobra.Command {
	var outputFormat string
	var selector string
	var showLabels bool
	var watch bool
	var watchOnly bool

	cmd := &cobra.Command{
		Use:   "get [TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...]",
		Short: "Display one or many resources across all managed clusters",
		Long: `Get resources from all managed clusters and display them in a unified view.
Supports all resource types that kubectl get supports.

The output includes cluster context information to help identify which
cluster each resource belongs to.`,
		Example: `# List all pods in all managed clusters
kubectl multi get pods

# List all nodes in all managed clusters
kubectl multi get nodes

# List deployments in specific namespace across all clusters
kubectl multi get deployments -n production

# List all resources in all namespaces across all clusters
kubectl multi get all -A

# Get pods with labels
kubectl multi get pods -l app=nginx

# Get specific pod across all clusters
kubectl multi get pod nginx-pod

# Get services with wide output
kubectl multi get services -o wide`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("resource type must be specified")
			}

			kubeconfig, remoteCtx, _, namespace, allNamespaces := GetGlobalFlags()
			return handleGetCommand(args, outputFormat, selector, showLabels, watch, watchOnly, kubeconfig, remoteCtx, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format (json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...)")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "selector (label query) to filter on")
	cmd.Flags().BoolVar(&showLabels, "show-labels", false, "show all labels as the last column")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for changes to the requested object(s)")
	cmd.Flags().BoolVar(&watchOnly, "watch-only", false, "watch for changes to the requested object(s), without listing/getting first")

	// Set custom help function
	cmd.SetHelpFunc(getHelpFunc)

	return cmd
}

func handleGetCommand(args []string, outputFormat, selector string, showLabels, watch, watchOnly bool, kubeconfig, remoteCtx, namespace string, allNamespaces bool) error {
	resourceType := args[0]
	resourceName := ""
	if len(args) > 1 {
		resourceName = args[1]
	}

	// For watch operations, we don't support multi-cluster watch yet
	if watch || watchOnly {
		return fmt.Errorf("watch operations are not supported in multi-cluster mode")
	}

	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteCtx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	tw := tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	defer tw.Flush()

	// Handle different resource types
	switch strings.ToLower(resourceType) {

	case "ingresses", "ingress", "ing":
		return handleIngressesGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "jobs", "job":
		return handleJobsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "cronjobs", "cronjob", "cj":
		return handleCronJobsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "all":
		return handleAllGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "nodes", "node", "no":
		return handleNodesGet(tw, clusters, resourceName, selector, showLabels, outputFormat)
	case "pods", "pod", "po":
		return handlePodsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "services", "service", "svc":
		return handleServicesGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "deployments", "deployment", "deploy":
		return handleDeploymentsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "replicasets", "replicaset", "rs":
		return handleReplicaSetsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "daemonsets", "daemonset", "ds":
		return handleDaemonSetsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "namespaces", "namespace", "ns":
		return handleNamespacesGet(tw, clusters, resourceName, selector, showLabels, outputFormat)
	case "configmaps", "configmap", "cm":
		return handleConfigMapsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "statefulsets", "statefulset", "sts":
		return handleStatefulSetsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "secrets", "secret":
		return handleSecretsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "persistentvolumes", "persistentvolume", "pv":
		return handlePVGet(tw, clusters, resourceName, selector, showLabels, outputFormat)
	case "persistentvolumeclaims", "persistentvolumeclaim", "pvc":
		return handlePVCGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	case "events", "event", "ev":
		return handleEventsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	default:
		return handleGenericGet(tw, clusters, resourceType, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces)
	}
}

func handleIngressesGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		ingresses, err := clusterInfo.Client.NetworkingV1().Ingresses(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list ingresses in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(ingresses.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tHOSTS\tADDRESS\tPORTS\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tHOSTS\tADDRESS\tPORTS\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tHOSTS\tADDRESS\tPORTS\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tHOSTS\tADDRESS\tPORTS\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, ing := range ingresses.Items {
			if resourceName != "" && ing.Name != resourceName {
				continue
			}

			// Gather hosts
			var hosts []string
			for _, rule := range ing.Spec.Rules {
				if rule.Host != "" {
					hosts = append(hosts, rule.Host)
				}
			}
			hostsStr := strings.Join(hosts, ",")
			if hostsStr == "" {
				hostsStr = "<none>"
			}

			// Gather address
			address := "<none>"
			if len(ing.Status.LoadBalancer.Ingress) > 0 {
				var addrs []string
				for _, lb := range ing.Status.LoadBalancer.Ingress {
					if lb.IP != "" {
						addrs = append(addrs, lb.IP)
					} else if lb.Hostname != "" {
						addrs = append(addrs, lb.Hostname)
					}
				}
				if len(addrs) > 0 {
					address = strings.Join(addrs, ",")
				}
			}

			// Gather ports
			var ports []string
			for _, rule := range ing.Spec.Rules {
				if rule.HTTP != nil {
					for range rule.HTTP.Paths {
						// Ingress does not specify port directly; default is 80
						ports = append(ports, "80")
					}
				}
			}
			// If TLS is specified, port 443 is implied
			if len(ing.Spec.TLS) > 0 {
				ports = append(ports, "443")
			}
			// Deduplicate ports
			portMap := make(map[string]struct{})
			for _, p := range ports {
				portMap[p] = struct{}{}
			}
			var portList []string
			for p := range portMap {
				portList = append(portList, p)
			}
			portsStr := strings.Join(portList, ",")
			if portsStr == "" {
				portsStr = "<none>"
			}

			age := duration.HumanDuration(time.Since(ing.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(ing.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, ing.Namespace, ing.Name, hostsStr, address, portsStr, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, ing.Namespace, ing.Name, hostsStr, address, portsStr, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(ing.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, ing.Name, hostsStr, address, portsStr, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, ing.Name, hostsStr, address, portsStr, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleJobsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		jobs, err := clusterInfo.Client.BatchV1().Jobs(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list jobs in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(jobs.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when items len is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tCOMPLETIONS\tDURATION\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tCOMPLETIONS\tDURATION\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tCOMPLETIONS\tDURATION\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tCOMPLETIONS\tDURATION\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, job := range jobs.Items {
			if resourceName != "" && job.Name != resourceName {
				continue
			}

			// Calculate completions
			var completions string
			if job.Spec.Completions != nil {
				completions = fmt.Sprintf("%d/%d", job.Status.Succeeded, *job.Spec.Completions)
			} else {
				completions = fmt.Sprintf("%d/1", job.Status.Succeeded)
			}

			// Calculate duration
			var jobDuration string
			if job.Status.StartTime != nil {
				if job.Status.CompletionTime != nil {
					jobDuration = duration.HumanDuration(job.Status.CompletionTime.Sub(job.Status.StartTime.Time))
				} else {
					jobDuration = duration.HumanDuration(time.Since(job.Status.StartTime.Time))
				}
			} else {
				jobDuration = "<unknown>"
			}

			// Calculate age
			age := duration.HumanDuration(time.Since(job.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(job.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, job.Namespace, job.Name, completions, jobDuration, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, job.Namespace, job.Name, completions, jobDuration, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(job.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, job.Name, completions, jobDuration, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, job.Name, completions, jobDuration, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}

	return nil
}

func handleAllGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	fmt.Println("==> Pods")
	if err := handlePodsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Services")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleServicesGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Deployments")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleDeploymentsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Jobs")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleJobsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> CronJobs")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleCronJobsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Nodes")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleNodesGet(tw, clusters, resourceName, selector, showLabels, outputFormat); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> ReplicaSets")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleReplicaSetsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> DaemonSets")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleDaemonSetsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Namespaces")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleNamespacesGet(tw, clusters, resourceName, selector, showLabels, outputFormat); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> ConfigMaps")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleConfigMapsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> StatefulSets")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleStatefulSetsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Secrets")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleSecretsGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> PersistentVolumes")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handlePVGet(tw, clusters, resourceName, selector, showLabels, outputFormat); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> PersistentVolumeClaims")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handlePVCGet(tw, clusters, resourceName, selector, showLabels, outputFormat, namespace, allNamespaces); err != nil {
		return err
	}
	tw.Flush()

	return nil
}
func handleNodesGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error {
	// Print header only once at the top
	if showLabels {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tROLES\tAGE\tVERSION\tLABELS\n")
	} else {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tROLES\tAGE\tVERSION\n")
	}

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		nodes, err := clusterInfo.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list nodes in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, node := range nodes.Items {
			if resourceName != "" && node.Name != resourceName {
				continue
			}

			status := util.GetNodeStatus(node)
			role := util.GetNodeRole(node)
			age := duration.HumanDuration(time.Since(node.CreationTimestamp.Time))
			version := node.Status.NodeInfo.KubeletVersion

			if showLabels {
				labels := util.FormatLabels(node.Labels)
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, node.Name, status, role, age, version, labels)
			} else {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, node.Name, status, role, age, version)
			}
		}
	}
	return nil
}

func handlePodsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		pods, err := clusterInfo.Client.CoreV1().Pods(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list pods in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(pods.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, pod := range pods.Items {
			if resourceName != "" && pod.Name != resourceName {
				continue
			}

			ready := fmt.Sprintf("%d/%d", util.GetPodReadyContainers(&pod), len(pod.Spec.Containers))
			status := string(pod.Status.Phase)
			restarts := util.GetPodRestarts(&pod)
			age := duration.HumanDuration(time.Since(pod.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(pod.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, pod.Namespace, pod.Name, ready, status, restarts, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, pod.Namespace, pod.Name, ready, status, restarts, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(pod.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, pod.Name, ready, status, restarts, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, pod.Name, ready, status, restarts, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleServicesGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		services, err := clusterInfo.Client.CoreV1().Services(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list services in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(services.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\n")
				}

			}
			isHeaderPrint = true
		}

		for _, svc := range services.Items {
			if resourceName != "" && svc.Name != resourceName {
				continue
			}

			svcType := string(svc.Spec.Type)
			clusterIP := svc.Spec.ClusterIP
			externalIP := util.GetServiceExternalIP(&svc)
			ports := util.GetServicePorts(&svc)
			age := duration.HumanDuration(time.Since(svc.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(svc.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, svc.Namespace, svc.Name, svcType, clusterIP, externalIP, ports, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, svc.Namespace, svc.Name, svcType, clusterIP, externalIP, ports, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(svc.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, svc.Name, svcType, clusterIP, externalIP, ports, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, svc.Name, svcType, clusterIP, externalIP, ports, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}

	return nil
}

func handleDeploymentsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		deployments, err := clusterInfo.Client.AppsV1().Deployments(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list deployments in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(deployments.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, deploy := range deployments.Items {
			if resourceName != "" && deploy.Name != resourceName {
				continue
			}

			var replicas int32 = 0
			if deploy.Spec.Replicas != nil {
				replicas = *deploy.Spec.Replicas
			}
			ready := fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, replicas)
			upToDate := fmt.Sprintf("%d", deploy.Status.UpdatedReplicas)
			available := fmt.Sprintf("%d", deploy.Status.AvailableReplicas)
			age := duration.HumanDuration(time.Since(deploy.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(deploy.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, deploy.Namespace, deploy.Name, ready, upToDate, available, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, deploy.Namespace, deploy.Name, ready, upToDate, available, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(deploy.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, deploy.Name, ready, upToDate, available, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, deploy.Name, ready, upToDate, available, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}

	return nil
}

func handleNamespacesGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error {
	// Print header only once at the top
	if showLabels {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tAGE\tLABELS\n")
	} else {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tAGE\n")
	}

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		namespaces, err := clusterInfo.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list namespaces in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, ns := range namespaces.Items {
			if resourceName != "" && ns.Name != resourceName {
				continue
			}

			status := string(ns.Status.Phase)
			age := duration.HumanDuration(time.Since(ns.CreationTimestamp.Time))

			if showLabels {
				labels := util.FormatLabels(ns.Labels)
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, ns.Name, status, age, labels)
			} else {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					clusterInfo.Name, ns.Name, status, age)
			}
		}
	}
	return nil
}

func handleConfigMapsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		configMaps, err := clusterInfo.Client.CoreV1().ConfigMaps(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list configmaps in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(configMaps.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDATA\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDATA\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tDATA\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tDATA\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, cm := range configMaps.Items {
			if resourceName != "" && cm.Name != resourceName {
				continue
			}

			dataCount := len(cm.Data) + len(cm.BinaryData)
			age := duration.HumanDuration(time.Since(cm.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(cm.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, cm.Namespace, cm.Name, dataCount, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, cm.Namespace, cm.Name, dataCount, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(cm.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, cm.Name, dataCount, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
						clusterInfo.Name, cm.Name, dataCount, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleSecretsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		secrets, err := clusterInfo.Client.CoreV1().Secrets(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list secrets in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(secrets.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tDATA\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tDATA\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tDATA\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tDATA\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, secret := range secrets.Items {
			if resourceName != "" && secret.Name != resourceName {
				continue
			}

			secretType := string(secret.Type)
			dataCount := len(secret.Data)
			age := duration.HumanDuration(time.Since(secret.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(secret.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, secret.Namespace, secret.Name, secretType, dataCount, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, secret.Namespace, secret.Name, secretType, dataCount, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(secret.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, secret.Name, secretType, dataCount, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, secret.Name, secretType, dataCount, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}

	return nil
}

func handlePVGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat string) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		pvs, err := clusterInfo.Client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list persistent volumes in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(pvs.Items) > 0 && !isHeaderPrint {
			if showLabels {
				fmt.Fprintf(tw, "CLUSTER\tNAME\tCAPACITY\tACCESS MODES\tRECLAIM POLICY\tSTATUS\tCLAIM\tSTORAGE CLASS\tREASON\tAGE\tLABELS\n")
			} else {
				fmt.Fprintf(tw, "CLUSTER\tNAME\tCAPACITY\tACCESS MODES\tRECLAIM POLICY\tSTATUS\tCLAIM\tSTORAGE CLASS\tREASON\tAGE\n")
			}
			isHeaderPrint = true
		}

		for _, pv := range pvs.Items {
			if resourceName != "" && pv.Name != resourceName {
				continue
			}

			capacity := util.GetPVCapacity(&pv)
			accessModes := util.GetPVAccessModes(&pv)
			reclaimPolicy := string(pv.Spec.PersistentVolumeReclaimPolicy)
			status := string(pv.Status.Phase)
			claim := util.GetPVClaim(&pv)
			storageClass := util.GetPVStorageClass(&pv)
			reason := pv.Status.Reason
			age := duration.HumanDuration(time.Since(pv.CreationTimestamp.Time))

			if showLabels {
				labels := util.FormatLabels(pv.Labels)
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, pv.Name, capacity, accessModes, reclaimPolicy, status, claim, storageClass, reason, age, labels)
			} else {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, pv.Name, capacity, accessModes, reclaimPolicy, status, claim, storageClass, reason, age)
			}
		}
	}

	if !isHeaderPrint {
		fmt.Fprintf(tw, "No resources found\n")
	}

	return nil
}

func handlePVCGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		pvcs, err := clusterInfo.Client.CoreV1().PersistentVolumeClaims(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list persistent volume claims in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(pvcs.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, pvc := range pvcs.Items {
			if resourceName != "" && pvc.Name != resourceName {
				continue
			}

			status := string(pvc.Status.Phase)
			volume := pvc.Spec.VolumeName
			capacity := util.GetPVCCapacity(&pvc)
			accessModes := util.GetPVCAccessModes(&pvc)
			storageClass := util.GetPVCStorageClass(&pvc)
			age := duration.HumanDuration(time.Since(pvc.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(pvc.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, pvc.Namespace, pvc.Name, status, volume, capacity, accessModes, storageClass, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, pvc.Namespace, pvc.Name, status, volume, capacity, accessModes, storageClass, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(pvc.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, pvc.Name, status, volume, capacity, accessModes, storageClass, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, pvc.Name, status, volume, capacity, accessModes, storageClass, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleGenericGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceType, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.DynamicClient == nil {
			continue
		}

		// Try to discover the resource
		gvr, isNamespaced, err := util.DiscoverGVR(clusterInfo.DiscoveryClient, resourceType)
		if err != nil {
			fmt.Printf("Warning: failed to discover resource %s in cluster %s: %v\n", resourceType, clusterInfo.Name, err)
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		var list *unstructured.UnstructuredList

		if isNamespaced && !allNamespaces && targetNS != "" {
			list, err = clusterInfo.DynamicClient.Resource(gvr).Namespace(targetNS).List(context.TODO(), metav1.ListOptions{
				LabelSelector: selector,
			})
		} else {
			list, err = clusterInfo.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{
				LabelSelector: selector,
			})
		}

		if err != nil {
			fmt.Printf("Warning: failed to list %s in cluster %s: %v\n", resourceType, clusterInfo.Name, err)
			continue
		}

		if len(list.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, item := range list.Items {
			if resourceName != "" && item.GetName() != resourceName {
				continue
			}

			age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))

			if isNamespaced && allNamespaces {
				if showLabels {
					labels := util.FormatLabels(item.GetLabels())
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, item.GetNamespace(), item.GetName(), age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
						clusterInfo.Name, item.GetNamespace(), item.GetName(), age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(item.GetLabels())
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
						clusterInfo.Name, item.GetName(), age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\n",
						clusterInfo.Name, item.GetName(), age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {

			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}

	return nil
}

func handleReplicaSetsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		replicaSets, err := clusterInfo.Client.AppsV1().ReplicaSets(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list replicasets in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(replicaSets.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDESIRED\tCURRENT\tREADY\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDESIRED\tCURRENT\tREADY\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tDESIRED\tCURRENT\tREADY\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tDESIRED\tCURRENT\tREADY\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, rs := range replicaSets.Items {
			if resourceName != "" && rs.Name != resourceName {
				continue
			}

			var desired int32 = 0
			if rs.Spec.Replicas != nil {
				desired = *rs.Spec.Replicas
			}
			current := rs.Status.Replicas
			ready := rs.Status.ReadyReplicas
			age := duration.HumanDuration(time.Since(rs.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(rs.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
						clusterInfo.Name, rs.Namespace, rs.Name, desired, current, ready, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\t%s\n",
						clusterInfo.Name, rs.Namespace, rs.Name, desired, current, ready, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(rs.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
						clusterInfo.Name, rs.Name, desired, current, ready, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%s\n",
						clusterInfo.Name, rs.Name, desired, current, ready, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleStatefulSetsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		statefulSets, err := clusterInfo.Client.AppsV1().StatefulSets(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list statefulsets in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(statefulSets.Items) > 0 && !isHeaderPrint {
			// Print header only once at top when any items is greater than 0.
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, sts := range statefulSets.Items {
			if resourceName != "" && sts.Name != resourceName {
				continue
			}

			var replicas int32 = 0
			if sts.Spec.Replicas != nil {
				replicas = *sts.Spec.Replicas
			}
			ready := fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, replicas)
			age := duration.HumanDuration(time.Since(sts.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(sts.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, sts.Namespace, sts.Name, ready, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, sts.Namespace, sts.Name, ready, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(sts.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, sts.Name, ready, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
						clusterInfo.Name, sts.Name, ready, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleDaemonSetsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		daemonSets, err := clusterInfo.Client.AppsV1().DaemonSets(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list daemonsets in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(daemonSets.Items) > 0 && !isHeaderPrint {
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDESIRED\tCURRENT\tREADY\tUP-TO-DATE\tAVAILABLE\tNODE SELECTOR\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDESIRED\tCURRENT\tREADY\tUP-TO-DATE\tAVAILABLE\tNODE SELECTOR\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tDESIRED\tCURRENT\tREADY\tUP-TO-DATE\tAVAILABLE\tNODE SELECTOR\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tDESIRED\tCURRENT\tREADY\tUP-TO-DATE\tAVAILABLE\tNODE SELECTOR\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, ds := range daemonSets.Items {
			if resourceName != "" && ds.Name != resourceName {
				continue
			}

			desired := ds.Status.DesiredNumberScheduled
			current := ds.Status.CurrentNumberScheduled
			ready := ds.Status.NumberReady
			upToDate := ds.Status.UpdatedNumberScheduled
			available := ds.Status.NumberAvailable

			// Format node selector
			nodeSelector := "<none>"
			if len(ds.Spec.Template.Spec.NodeSelector) > 0 {
				var selectors []string
				for k, v := range ds.Spec.Template.Spec.NodeSelector {
					selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
				}
				nodeSelector = strings.Join(selectors, ",")
			}

			age := duration.HumanDuration(time.Since(ds.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(ds.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
						clusterInfo.Name, ds.Namespace, ds.Name, desired, current, ready, upToDate, available, nodeSelector, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\n",
						clusterInfo.Name, ds.Namespace, ds.Name, desired, current, ready, upToDate, available, nodeSelector, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(ds.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
						clusterInfo.Name, ds.Name, desired, current, ready, upToDate, available, nodeSelector, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\n",
						clusterInfo.Name, ds.Name, desired, current, ready, upToDate, available, nodeSelector, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleCronJobsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	isHeaderPrint := false

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		cronJobs, err := clusterInfo.Client.BatchV1().CronJobs(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list cronjobs in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		if len(cronJobs.Items) > 0 && !isHeaderPrint {
			if allNamespaces {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tSCHEDULE\tSUSPEND\tACTIVE\tLAST SCHEDULE\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tSCHEDULE\tSUSPEND\tACTIVE\tLAST SCHEDULE\tAGE\n")
				}
			} else {
				if showLabels {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tSCHEDULE\tSUSPEND\tACTIVE\tLAST SCHEDULE\tAGE\tLABELS\n")
				} else {
					fmt.Fprintf(tw, "CLUSTER\tNAME\tSCHEDULE\tSUSPEND\tACTIVE\tLAST SCHEDULE\tAGE\n")
				}
			}
			isHeaderPrint = true
		}

		for _, cj := range cronJobs.Items {
			if resourceName != "" && cj.Name != resourceName {
				continue
			}

			schedule := cj.Spec.Schedule

			suspend := "False"
			if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
				suspend = "True"
			}

			active := len(cj.Status.Active)

			lastSchedule := "<none>"
			if cj.Status.LastScheduleTime != nil {
				lastSchedule = duration.HumanDuration(time.Since(cj.Status.LastScheduleTime.Time))
			}

			age := duration.HumanDuration(time.Since(cj.CreationTimestamp.Time))

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(cj.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
						clusterInfo.Name, cj.Namespace, cj.Name, schedule, suspend, active, lastSchedule, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, cj.Namespace, cj.Name, schedule, suspend, active, lastSchedule, age)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(cj.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
						clusterInfo.Name, cj.Name, schedule, suspend, active, lastSchedule, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, cj.Name, schedule, suspend, active, lastSchedule, age)
				}
			}
		}
	}

	if !isHeaderPrint {
		// print no resource found if isHeaderPrint is still false at this point
		if allNamespaces {
			fmt.Fprintf(tw, "No resource found.\n")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Fprintf(tw, "No resource found in %s namespace.\n", namespace)
		}
	}
	return nil
}

func handleEventsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, resourceName, selector string, showLabels bool, outputFormat, namespace string, allNamespaces bool) error {
	if allNamespaces {
		if showLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tLAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tLAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\n")
		}
	} else {
		if showLabels {
			fmt.Fprintf(tw, "CLUSTER\tLAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tLAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\n")
		}
	}

	for _, clusterInfo := range clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(namespace)
		if allNamespaces {
			targetNS = ""
		}

		events, err := clusterInfo.Client.CoreV1().Events(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list events in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, event := range events.Items {
			if resourceName != "" && event.Name != resourceName {
				continue
			}

			lastSeen := "<unknown>"
			if !event.LastTimestamp.IsZero() {
				lastSeen = duration.HumanDuration(time.Since(event.LastTimestamp.Time)) + " ago"
			} else if !event.FirstTimestamp.IsZero() {
				lastSeen = duration.HumanDuration(time.Since(event.FirstTimestamp.Time)) + " ago"
			}

			eventType := event.Type
			reason := event.Reason
			object := fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name)
			message := event.Message

			if allNamespaces {
				if showLabels {
					labels := util.FormatLabels(event.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, event.Namespace, lastSeen, eventType, reason, object, message, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, event.Namespace, lastSeen, eventType, reason, object, message)
				}
			} else {
				if showLabels {
					labels := util.FormatLabels(event.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, lastSeen, eventType, reason, object, message, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, lastSeen, eventType, reason, object, message)
				}
			}
		}
	}
	return nil
}
