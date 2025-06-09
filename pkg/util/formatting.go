package util

import (
	"fmt"
	"os"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

// GetOutputStream returns the output stream (stdout)
func GetOutputStream() *os.File {
	return os.Stdout
}

// GetNodeStatus returns the status of a node
func GetNodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

// GetNodeRole returns the role of a node
func GetNodeRole(node corev1.Node) string {
	for label := range node.Labels {
		const prefix = "node-role.kubernetes.io/"
		if strings.HasPrefix(label, prefix) {
			role := strings.TrimPrefix(label, prefix)
			if role != "" {
				return role
			}
		}
	}
	return "<none>"
}

// GetPodReadyContainers returns the number of ready containers in a pod
func GetPodReadyContainers(pod *corev1.Pod) int32 {
	var ready int32
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}

// GetPodRestarts returns the total number of restarts for all containers in a pod
func GetPodRestarts(pod *corev1.Pod) int32 {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

// GetServiceExternalIP returns the external IP of a service
func GetServiceExternalIP(svc *corev1.Service) string {
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ingress := svc.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			return ingress.IP
		}
		if ingress.Hostname != "" {
			return ingress.Hostname
		}
	}
	if len(svc.Spec.ExternalIPs) > 0 {
		return strings.Join(svc.Spec.ExternalIPs, ",")
	}
	return "<none>"
}

// GetServicePorts returns the ports of a service formatted as a string
func GetServicePorts(svc *corev1.Service) string {
	var ports []string
	for _, port := range svc.Spec.Ports {
		if port.NodePort != 0 {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
		}
	}
	if len(ports) == 0 {
		return "<none>"
	}
	return strings.Join(ports, ",")
}

// FormatLabels formats a map of labels as a string
func FormatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}

	var items []string
	for k, v := range labels {
		items = append(items, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(items)
	return strings.Join(items, ",")
}

// GetPVCapacity returns the capacity of a persistent volume
func GetPVCapacity(pv *corev1.PersistentVolume) string {
	if capacity, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
		return capacity.String()
	}
	return "<unknown>"
}

// GetPVAccessModes returns the access modes of a persistent volume
func GetPVAccessModes(pv *corev1.PersistentVolume) string {
	var modes []string
	for _, mode := range pv.Spec.AccessModes {
		switch mode {
		case corev1.ReadWriteOnce:
			modes = append(modes, "RWO")
		case corev1.ReadOnlyMany:
			modes = append(modes, "ROX")
		case corev1.ReadWriteMany:
			modes = append(modes, "RWX")
		case corev1.ReadWriteOncePod:
			modes = append(modes, "RWOP")
		default:
			modes = append(modes, string(mode))
		}
	}
	return strings.Join(modes, ",")
}

// GetPVClaim returns the claim name for a persistent volume
func GetPVClaim(pv *corev1.PersistentVolume) string {
	if pv.Spec.ClaimRef != nil {
		return fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}
	return "<none>"
}

// GetPVStorageClass returns the storage class of a persistent volume
func GetPVStorageClass(pv *corev1.PersistentVolume) string {
	if pv.Spec.StorageClassName != "" {
		return pv.Spec.StorageClassName
	}
	return "<none>"
}

// GetPVCCapacity returns the capacity of a persistent volume claim
func GetPVCCapacity(pvc *corev1.PersistentVolumeClaim) string {
	if capacity, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
		return capacity.String()
	}
	return "<unset>"
}

// GetPVCAccessModes returns the access modes of a persistent volume claim
func GetPVCAccessModes(pvc *corev1.PersistentVolumeClaim) string {
	var modes []string
	for _, mode := range pvc.Status.AccessModes {
		switch mode {
		case corev1.ReadWriteOnce:
			modes = append(modes, "RWO")
		case corev1.ReadOnlyMany:
			modes = append(modes, "ROX")
		case corev1.ReadWriteMany:
			modes = append(modes, "RWX")
		case corev1.ReadWriteOncePod:
			modes = append(modes, "RWOP")
		default:
			modes = append(modes, string(mode))
		}
	}
	return strings.Join(modes, ",")
}

// GetPVCStorageClass returns the storage class of a persistent volume claim
func GetPVCStorageClass(pvc *corev1.PersistentVolumeClaim) string {
	if pvc.Spec.StorageClassName != nil {
		return *pvc.Spec.StorageClassName
	}
	return "<none>"
}

// DiscoverGVR discovers the GroupVersionResource for a given resource type
func DiscoverGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (schema.GroupVersionResource, bool, error) {
	// Get all API resources
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return schema.GroupVersionResource{}, false, fmt.Errorf("failed to discover API resources: %v", err)
	}

	// Normalize the resource type (handle plurals and common aliases)
	normalizedType := normalizeResourceType(resourceType)

	// Search through all API resources
	for _, apiResourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			// Check if this matches our resource type
			if matchesResourceType(apiResource, normalizedType) {
				gvr := gv.WithResource(apiResource.Name)
				return gvr, apiResource.Namespaced, nil
			}
		}
	}

	// If not found, try some common defaults
	return getDefaultGVR(normalizedType), true, nil
}

// normalizeResourceType converts common resource type aliases to standard forms
func normalizeResourceType(resourceType string) string {
	aliases := map[string]string{
		"po":     "pods",
		"svc":    "services",
		"no":     "nodes",
		"ns":     "namespaces",
		"pv":     "persistentvolumes",
		"pvc":    "persistentvolumeclaims",
		"cm":     "configmaps",
		"deploy": "deployments",
		"rs":     "replicasets",
		"ds":     "daemonsets",
		"sts":    "statefulsets",
		"job":    "jobs",
		"cj":     "cronjobs",
		"ing":    "ingresses",
		"ep":     "endpoints",
		"sa":     "serviceaccounts",
	}

	if normalized, exists := aliases[strings.ToLower(resourceType)]; exists {
		return normalized
	}

	// Ensure it's lowercase and plural
	lower := strings.ToLower(resourceType)
	if !strings.HasSuffix(lower, "s") {
		lower += "s"
	}
	return lower
}

// matchesResourceType checks if an API resource matches the given resource type
func matchesResourceType(apiResource metav1.APIResource, resourceType string) bool {
	// Check exact match with name
	if strings.EqualFold(apiResource.Name, resourceType) {
		return true
	}

	// Check singular name
	if strings.EqualFold(apiResource.SingularName, resourceType) {
		return true
	}

	// Check short names
	for _, shortName := range apiResource.ShortNames {
		if strings.EqualFold(shortName, resourceType) {
			return true
		}
	}

	return false
}

// getDefaultGVR returns a default GVR for common resource types
func getDefaultGVR(resourceType string) schema.GroupVersionResource {
	defaults := map[string]schema.GroupVersionResource{
		"pods":                   {Group: "", Version: "v1", Resource: "pods"},
		"services":               {Group: "", Version: "v1", Resource: "services"},
		"nodes":                  {Group: "", Version: "v1", Resource: "nodes"},
		"namespaces":             {Group: "", Version: "v1", Resource: "namespaces"},
		"persistentvolumes":      {Group: "", Version: "v1", Resource: "persistentvolumes"},
		"persistentvolumeclaims": {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		"configmaps":             {Group: "", Version: "v1", Resource: "configmaps"},
		"secrets":                {Group: "", Version: "v1", Resource: "secrets"},
		"deployments":            {Group: "apps", Version: "v1", Resource: "deployments"},
		"replicasets":            {Group: "apps", Version: "v1", Resource: "replicasets"},
		"daemonsets":             {Group: "apps", Version: "v1", Resource: "daemonsets"},
		"statefulsets":           {Group: "apps", Version: "v1", Resource: "statefulsets"},
		"jobs":                   {Group: "batch", Version: "v1", Resource: "jobs"},
		"cronjobs":               {Group: "batch", Version: "v1", Resource: "cronjobs"},
		"ingresses":              {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		"endpoints":              {Group: "", Version: "v1", Resource: "endpoints"},
		"serviceaccounts":        {Group: "", Version: "v1", Resource: "serviceaccounts"},
	}

	if gvr, exists := defaults[resourceType]; exists {
		return gvr
	}

	// Default fallback
	return schema.GroupVersionResource{Group: "", Version: "v1", Resource: resourceType}
}
