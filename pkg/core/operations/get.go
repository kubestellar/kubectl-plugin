package operations

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/duration"

	"kubectl-multi/pkg/cluster"
	"kubectl-multi/pkg/util"
)

// GetOptions contains all the options for the get operation
type GetOptions struct {
	Clusters      []cluster.ClusterInfo
	ResourceType  string
	ResourceName  string
	Namespace     string
	AllNamespaces bool
	Selector      string
	ShowLabels    bool
	OutputFormat  string
}

// ExecuteGet performs the get operation across all clusters
func ExecuteGet(opts GetOptions) error {
	tw := tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	defer tw.Flush()

	// Handle different resource types
	switch strings.ToLower(opts.ResourceType) {
	case "jobs", "job":
		return handleJobsGet(tw, opts)
	case "all":
		return handleAllGet(tw, opts)
	case "nodes", "node", "no":
		return handleNodesGet(tw, opts)
	case "pods", "pod", "po":
		return handlePodsGet(tw, opts)
	case "services", "service", "svc":
		return handleServicesGet(tw, opts)
	case "deployments", "deployment", "deploy":
		return handleDeploymentsGet(tw, opts)
	case "namespaces", "namespace", "ns":
		return handleNamespacesGet(tw, opts)
	case "configmaps", "configmap", "cm":
		return handleConfigMapsGet(tw, opts)
	case "secrets", "secret":
		return handleSecretsGet(tw, opts)
	case "persistentvolumes", "persistentvolume", "pv":
		return handlePVGet(tw, opts)
	case "persistentvolumeclaims", "persistentvolumeclaim", "pvc":
		return handlePVCGet(tw, opts)
	default:
		return handleGenericGet(tw, opts)
	}
}

func handleJobsGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tCOMPLETIONS\tDURATION\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tCOMPLETIONS\tDURATION\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tCOMPLETIONS\tDURATION\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tCOMPLETIONS\tDURATION\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		jobs, err := clusterInfo.Client.BatchV1().Jobs(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list jobs in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, job := range jobs.Items {
			if opts.ResourceName != "" && job.Name != opts.ResourceName {
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

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(job.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, job.Namespace, job.Name, completions, jobDuration, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, job.Namespace, job.Name, completions, jobDuration, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handleAllGet(tw *tabwriter.Writer, opts GetOptions) error {
	fmt.Println("==> Pods")
	if err := handlePodsGet(tw, opts); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Services")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleServicesGet(tw, opts); err != nil {
		return err
	}
	tw.Flush()

	fmt.Println("\n==> Deployments")
	tw = tabwriter.NewWriter(util.GetOutputStream(), 0, 0, 2, ' ', 0)
	if err := handleDeploymentsGet(tw, opts); err != nil {
		return err
	}
	tw.Flush()
	return nil
}

func handleNodesGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.ShowLabels {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tROLES\tAGE\tVERSION\tLABELS\n")
	} else {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tROLES\tAGE\tVERSION\n")
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		nodes, err := clusterInfo.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list nodes in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, node := range nodes.Items {
			if opts.ResourceName != "" && node.Name != opts.ResourceName {
				continue
			}

			status := util.GetNodeStatus(node)
			role := util.GetNodeRole(node)
			age := duration.HumanDuration(time.Since(node.CreationTimestamp.Time))
			version := node.Status.NodeInfo.KubeletVersion

			if opts.ShowLabels {
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

func handlePodsGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		pods, err := clusterInfo.Client.CoreV1().Pods(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list pods in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, pod := range pods.Items {
			if opts.ResourceName != "" && pod.Name != opts.ResourceName {
				continue
			}

			ready := fmt.Sprintf("%d/%d", util.GetPodReadyContainers(&pod), len(pod.Spec.Containers))
			status := string(pod.Status.Phase)
			restarts := util.GetPodRestarts(&pod)
			age := duration.HumanDuration(time.Since(pod.CreationTimestamp.Time))

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(pod.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, pod.Namespace, pod.Name, ready, status, restarts, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, pod.Namespace, pod.Name, ready, status, restarts, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handleServicesGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		services, err := clusterInfo.Client.CoreV1().Services(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list services in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, svc := range services.Items {
			if opts.ResourceName != "" && svc.Name != opts.ResourceName {
				continue
			}

			svcType := string(svc.Spec.Type)
			clusterIP := svc.Spec.ClusterIP
			externalIP := util.GetServiceExternalIP(&svc)
			ports := util.GetServicePorts(&svc)
			age := duration.HumanDuration(time.Since(svc.CreationTimestamp.Time))

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(svc.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, svc.Namespace, svc.Name, svcType, clusterIP, externalIP, ports, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, svc.Namespace, svc.Name, svcType, clusterIP, externalIP, ports, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handleDeploymentsGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		deployments, err := clusterInfo.Client.AppsV1().Deployments(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list deployments in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, deploy := range deployments.Items {
			if opts.ResourceName != "" && deploy.Name != opts.ResourceName {
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

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(deploy.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, deploy.Namespace, deploy.Name, ready, upToDate, available, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, deploy.Namespace, deploy.Name, ready, upToDate, available, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handleNamespacesGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.ShowLabels {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tAGE\tLABELS\n")
	} else {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tAGE\n")
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		namespaces, err := clusterInfo.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list namespaces in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, ns := range namespaces.Items {
			if opts.ResourceName != "" && ns.Name != opts.ResourceName {
				continue
			}

			status := string(ns.Status.Phase)
			age := duration.HumanDuration(time.Since(ns.CreationTimestamp.Time))

			if opts.ShowLabels {
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

func handleConfigMapsGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDATA\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tDATA\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tDATA\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tDATA\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		configMaps, err := clusterInfo.Client.CoreV1().ConfigMaps(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list configmaps in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, cm := range configMaps.Items {
			if opts.ResourceName != "" && cm.Name != opts.ResourceName {
				continue
			}

			dataCount := len(cm.Data) + len(cm.BinaryData)
			age := duration.HumanDuration(time.Since(cm.CreationTimestamp.Time))

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(cm.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, cm.Namespace, cm.Name, dataCount, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, cm.Namespace, cm.Name, dataCount, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handleSecretsGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tDATA\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tTYPE\tDATA\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tDATA\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tTYPE\tDATA\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		secrets, err := clusterInfo.Client.CoreV1().Secrets(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list secrets in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, secret := range secrets.Items {
			if opts.ResourceName != "" && secret.Name != opts.ResourceName {
				continue
			}

			secretType := string(secret.Type)
			dataCount := len(secret.Data)
			age := duration.HumanDuration(time.Since(secret.CreationTimestamp.Time))

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(secret.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
						clusterInfo.Name, secret.Namespace, secret.Name, secretType, dataCount, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
						clusterInfo.Name, secret.Namespace, secret.Name, secretType, dataCount, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handlePVGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.ShowLabels {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tCAPACITY\tACCESS MODES\tRECLAIM POLICY\tSTATUS\tCLAIM\tSTORAGE CLASS\tREASON\tAGE\tLABELS\n")
	} else {
		fmt.Fprintf(tw, "CLUSTER\tNAME\tCAPACITY\tACCESS MODES\tRECLAIM POLICY\tSTATUS\tCLAIM\tSTORAGE CLASS\tREASON\tAGE\n")
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		pvs, err := clusterInfo.Client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list persistent volumes in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, pv := range pvs.Items {
			if opts.ResourceName != "" && pv.Name != opts.ResourceName {
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

			if opts.ShowLabels {
				labels := util.FormatLabels(pv.Labels)
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, pv.Name, capacity, accessModes, reclaimPolicy, status, claim, storageClass, reason, age, labels)
			} else {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					clusterInfo.Name, pv.Name, capacity, accessModes, reclaimPolicy, status, claim, storageClass, reason, age)
			}
		}
	}
	return nil
}

func handlePVCGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tSTATUS\tVOLUME\tCAPACITY\tACCESS MODES\tSTORAGE CLASS\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.Client == nil {
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		if opts.AllNamespaces {
			targetNS = ""
		}

		pvcs, err := clusterInfo.Client.CoreV1().PersistentVolumeClaims(targetNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: opts.Selector,
		})
		if err != nil {
			fmt.Printf("Warning: failed to list persistent volume claims in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}

		for _, pvc := range pvcs.Items {
			if opts.ResourceName != "" && pvc.Name != opts.ResourceName {
				continue
			}

			status := string(pvc.Status.Phase)
			volume := pvc.Spec.VolumeName
			capacity := util.GetPVCCapacity(&pvc)
			accessModes := util.GetPVCAccessModes(&pvc)
			storageClass := util.GetPVCStorageClass(&pvc)
			age := duration.HumanDuration(time.Since(pvc.CreationTimestamp.Time))

			if opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(pvc.Labels)
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, pvc.Namespace, pvc.Name, status, volume, capacity, accessModes, storageClass, age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, pvc.Namespace, pvc.Name, status, volume, capacity, accessModes, storageClass, age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}

func handleGenericGet(tw *tabwriter.Writer, opts GetOptions) error {
	// Print header only once at the top for generic resources
	if opts.AllNamespaces {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAMESPACE\tNAME\tAGE\n")
		}
	} else {
		if opts.ShowLabels {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tAGE\tLABELS\n")
		} else {
			fmt.Fprintf(tw, "CLUSTER\tNAME\tAGE\n")
		}
	}

	for _, clusterInfo := range opts.Clusters {
		if clusterInfo.DynamicClient == nil {
			continue
		}

		// Try to discover the resource
		gvr, isNamespaced, err := util.DiscoverGVR(clusterInfo.DiscoveryClient, opts.ResourceType)
		if err != nil {
			fmt.Printf("Warning: failed to discover resource %s in cluster %s: %v\n", opts.ResourceType, clusterInfo.Name, err)
			continue
		}

		targetNS := cluster.GetTargetNamespace(opts.Namespace)
		var list *unstructured.UnstructuredList

		if isNamespaced && !opts.AllNamespaces && targetNS != "" {
			list, err = clusterInfo.DynamicClient.Resource(gvr).Namespace(targetNS).List(context.TODO(), metav1.ListOptions{
				LabelSelector: opts.Selector,
			})
		} else {
			list, err = clusterInfo.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{
				LabelSelector: opts.Selector,
			})
		}

		if err != nil {
			fmt.Printf("Warning: failed to list %s in cluster %s: %v\n", opts.ResourceType, clusterInfo.Name, err)
			continue
		}

		for _, item := range list.Items {
			if opts.ResourceName != "" && item.GetName() != opts.ResourceName {
				continue
			}

			age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))

			if isNamespaced && opts.AllNamespaces {
				if opts.ShowLabels {
					labels := util.FormatLabels(item.GetLabels())
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
						clusterInfo.Name, item.GetNamespace(), item.GetName(), age, labels)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
						clusterInfo.Name, item.GetNamespace(), item.GetName(), age)
				}
			} else {
				if opts.ShowLabels {
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
	return nil
}