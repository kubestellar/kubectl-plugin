package util

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

// ColumnDefinition represents a column header and how to extract its value
type ColumnDefinition struct {
	Name         string
	JSONPath     string
	DefaultValue string
}

// GetResourceColumns returns the appropriate column definitions for a resource type
func GetResourceColumns(resourceType string) []ColumnDefinition {
	columns := map[string][]ColumnDefinition{
		"pods": {
			{Name: "READY", JSONPath: "status.containerStatuses", DefaultValue: "<unknown>"},
			{Name: "STATUS", JSONPath: "status.phase", DefaultValue: "<unknown>"},
			{Name: "RESTARTS", JSONPath: "status.containerStatuses", DefaultValue: "<unknown>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"services": {
			{Name: "TYPE", JSONPath: "spec.type", DefaultValue: "ClusterIP"},
			{Name: "CLUSTER-IP", JSONPath: "spec.clusterIP", DefaultValue: "<none>"},
			{Name: "EXTERNAL-IP", JSONPath: "status.loadBalancer.ingress", DefaultValue: "<none>"},
			{Name: "PORT(S)", JSONPath: "spec.ports", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"deployments": {
			{Name: "READY", JSONPath: "status.readyReplicas", DefaultValue: "0"},
			{Name: "UP-TO-DATE", JSONPath: "status.updatedReplicas", DefaultValue: "0"},
			{Name: "AVAILABLE", JSONPath: "status.availableReplicas", DefaultValue: "0"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"replicasets": {
			{Name: "DESIRED", JSONPath: "spec.replicas", DefaultValue: "0"},
			{Name: "CURRENT", JSONPath: "status.replicas", DefaultValue: "0"},
			{Name: "READY", JSONPath: "status.readyReplicas", DefaultValue: "0"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"daemonsets": {
			{Name: "DESIRED", JSONPath: "status.desiredNumberScheduled", DefaultValue: "0"},
			{Name: "CURRENT", JSONPath: "status.currentNumberScheduled", DefaultValue: "0"},
			{Name: "READY", JSONPath: "status.numberReady", DefaultValue: "0"},
			{Name: "UP-TO-DATE", JSONPath: "status.updatedNumberScheduled", DefaultValue: "0"},
			{Name: "AVAILABLE", JSONPath: "status.numberAvailable", DefaultValue: "0"},
			{Name: "NODE SELECTOR", JSONPath: "spec.template.spec.nodeSelector", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"statefulsets": {
			{Name: "READY", JSONPath: "status.readyReplicas", DefaultValue: "0"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"jobs": {
			{Name: "COMPLETIONS", JSONPath: "status", DefaultValue: "<none>"},
			{Name: "DURATION", JSONPath: "status.startTime", DefaultValue: "<unknown>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"cronjobs": {
			{Name: "SCHEDULE", JSONPath: "spec.schedule", DefaultValue: "<none>"},
			{Name: "SUSPEND", JSONPath: "spec.suspend", DefaultValue: "False"},
			{Name: "ACTIVE", JSONPath: "status.active", DefaultValue: "0"},
			{Name: "LAST SCHEDULE", JSONPath: "status.lastScheduleTime", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"configmaps": {
			{Name: "DATA", JSONPath: "data", DefaultValue: "0"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"secrets": {
			{Name: "TYPE", JSONPath: "type", DefaultValue: "Opaque"},
			{Name: "DATA", JSONPath: "data", DefaultValue: "0"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"persistentvolumes": {
			{Name: "CAPACITY", JSONPath: "spec.capacity", DefaultValue: "<unknown>"},
			{Name: "ACCESS MODES", JSONPath: "spec.accessModes", DefaultValue: "<unknown>"},
			{Name: "RECLAIM POLICY", JSONPath: "spec.persistentVolumeReclaimPolicy", DefaultValue: "<unknown>"},
			{Name: "STATUS", JSONPath: "status.phase", DefaultValue: "<unknown>"},
			{Name: "CLAIM", JSONPath: "spec.claimRef", DefaultValue: "<none>"},
			{Name: "STORAGE CLASS", JSONPath: "spec.storageClassName", DefaultValue: "<none>"},
			{Name: "REASON", JSONPath: "status.reason", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"persistentvolumeclaims": {
			{Name: "STATUS", JSONPath: "status.phase", DefaultValue: "<unknown>"},
			{Name: "VOLUME", JSONPath: "spec.volumeName", DefaultValue: "<none>"},
			{Name: "CAPACITY", JSONPath: "status.capacity", DefaultValue: "<unknown>"},
			{Name: "ACCESS MODES", JSONPath: "status.accessModes", DefaultValue: "<unknown>"},
			{Name: "STORAGE CLASS", JSONPath: "spec.storageClassName", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"ingresses": {
			{Name: "HOSTS", JSONPath: "spec.rules", DefaultValue: "<none>"},
			{Name: "ADDRESS", JSONPath: "status.loadBalancer.ingress", DefaultValue: "<none>"},
			{Name: "PORTS", JSONPath: "spec.rules", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"endpoints": {
			{Name: "ENDPOINTS", JSONPath: "subsets", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"serviceaccounts": {
			{Name: "SECRETS", JSONPath: "secrets", DefaultValue: "0"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"resourcequotas": {
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
			{Name: "HARD", JSONPath: "status.hard", DefaultValue: "<none>"},
			{Name: "USED", JSONPath: "status.used", DefaultValue: "<none>"},
		},
		"limitranges": {
			{Name: "CREATED AT", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"networkpolicies": {
			{Name: "POD-SELECTOR", JSONPath: "spec.podSelector", DefaultValue: "<none>"},
			{Name: "POLICY-TYPES", JSONPath: "spec.policyTypes", DefaultValue: "<none>"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"roles": {
			{Name: "CREATED-AT", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"storageclasses": {
			{Name: "PROVISIONER", JSONPath: "provisioner", DefaultValue: "<none>"},
			{Name: "RECLAIMPOLICY", JSONPath: "reclaimPolicy", DefaultValue: "Delete"},
			{Name: "VOLUMEBINDINGMODE", JSONPath: "volumeBindingMode", DefaultValue: "Immediate"},
			{Name: "ALLOWVOLUMEEXPANSION", JSONPath: "allowVolumeExpansion", DefaultValue: "false"},
			{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
		},
		"events": {
			{Name: "LAST SEEN", JSONPath: "lastTimestamp", DefaultValue: "<unknown>"},
			{Name: "TYPE", JSONPath: "type", DefaultValue: "<unknown>"},
			{Name: "REASON", JSONPath: "reason", DefaultValue: "<unknown>"},
			{Name: "OBJECT", JSONPath: "", DefaultValue: "<unknown>"},
			{Name: "MESSAGE", JSONPath: "message", DefaultValue: "<unknown>"},
		},
	}

	if cols, exists := columns[strings.ToLower(resourceType)]; exists {
		return cols
	}

	// Default fallback - just show AGE
	return []ColumnDefinition{
		{Name: "AGE", JSONPath: "metadata.creationTimestamp", DefaultValue: "<unknown>"},
	}
}

// ExtractColumnValue extracts a column value from an unstructured object using JSONPath
func ExtractColumnValue(obj *unstructured.Unstructured, column ColumnDefinition) string {
	if column.JSONPath == "" {
		return column.DefaultValue
	}

	// Special handling for complex fields that need custom formatting
	switch column.Name {
	case "READY":
		return extractReadyValue(obj)
	case "STATUS":
		return extractStatusValue(obj)
	case "RESTARTS":
		return extractRestartsValue(obj)
	case "TYPE":
		return extractTypeValue(obj)
	case "CLUSTER-IP":
		return extractClusterIPValue(obj)
	case "EXTERNAL-IP":
		return extractExternalIPValue(obj)
	case "PORT(S)":
		return extractPortsValue(obj)
	case "UP-TO-DATE":
		return extractUpToDateValue(obj)
	case "AVAILABLE":
		return extractAvailableValue(obj)
	case "DESIRED":
		return extractDesiredValue(obj)
	case "CURRENT":
		return extractCurrentValue(obj)
	case "COMPLETIONS":
		return extractCompletionsValue(obj)
	case "DURATION":
		return extractDurationValue(obj)
	case "DATA":
		return extractDataCountValue(obj)
	case "CAPACITY":
		return extractCapacityValue(obj)
	case "ACCESS MODES":
		return extractAccessModesValue(obj)
	case "CLAIM":
		return extractClaimValue(obj)
	case "STORAGE CLASS":
		return extractStorageClassValue(obj)
	case "HOSTS":
		return extractHostsValue(obj)
	case "ADDRESS":
		return extractAddressValue(obj)
	case "PORTS":
		return extractIngressPortsValue(obj)
	case "ENDPOINTS":
		return extractEndpointsValue(obj)
	case "SECRETS":
		return extractSecretsValue(obj)
	case "HARD":
		return extractHardValue(obj)
	case "USED":
		return extractUsedValue(obj)
	case "POD-SELECTOR":
		return extractPodSelectorValue(obj)
	case "POLICY-TYPES":
		return extractPolicyTypesValue(obj)
	case "LAST SEEN":
		return extractLastSeenValue(obj)
	case "OBJECT":
		return extractObjectValue(obj)
	case "NODE SELECTOR":
		return extractNodeSelectorValue(obj)
	case "ALLOWVOLUMEEXPANSION":
		return extractAllowVolumeExpansionValue(obj)
	case "SCHEDULE":
		return extractScheduleValue(obj)
	case "SUSPEND":
		return extractSuspendValue(obj)
	case "ACTIVE":
		return extractActiveValue(obj)
	case "LAST SCHEDULE":
		return extractLastScheduleValue(obj)
	case "CREATED AT", "CREATED-AT":
		return extractCreatedAtValue(obj)
	}

	// For simple JSONPath extractions
	val, found, err := unstructured.NestedString(obj.Object, strings.Split(column.JSONPath, ".")...)
	if err != nil || !found {
		return column.DefaultValue
	}
	return val
}

// Helper functions for extracting complex column values

func extractReadyValue(obj *unstructured.Unstructured) string {
	containerStatuses, found, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses")
	if !found {
		return "0/0"
	}

	readyCount := 0
	totalCount := len(containerStatuses)

	for _, status := range containerStatuses {
		if statusMap, ok := status.(map[string]interface{}); ok {
			if ready, exists := statusMap["ready"]; exists {
				if readyBool, ok := ready.(bool); ok && readyBool {
					readyCount++
				}
			}
		}
	}

	return fmt.Sprintf("%d/%d", readyCount, totalCount)
}

func extractStatusValue(obj *unstructured.Unstructured) string {
	phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase")
	if !found {
		return "<unknown>"
	}
	return phase
}

func extractRestartsValue(obj *unstructured.Unstructured) string {
	containerStatuses, found, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses")
	if !found {
		return "0"
	}

	totalRestarts := 0
	for _, status := range containerStatuses {
		if statusMap, ok := status.(map[string]interface{}); ok {
			if restartCount, exists := statusMap["restartCount"]; exists {
				if count, ok := restartCount.(int64); ok {
					totalRestarts += int(count)
				}
			}
		}
	}

	return fmt.Sprintf("%d", totalRestarts)
}

func extractTypeValue(obj *unstructured.Unstructured) string {
	svcType, found, _ := unstructured.NestedString(obj.Object, "spec", "type")
	if !found {
		return "ClusterIP"
	}
	return svcType
}

func extractClusterIPValue(obj *unstructured.Unstructured) string {
	clusterIP, found, _ := unstructured.NestedString(obj.Object, "spec", "clusterIP")
	if !found {
		return "<none>"
	}
	return clusterIP
}

func extractExternalIPValue(obj *unstructured.Unstructured) string {
	ingress, found, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress")
	if !found || len(ingress) == 0 {
		return "<none>"
	}

	var ips []string
	for _, ing := range ingress {
		if ingMap, ok := ing.(map[string]interface{}); ok {
			if ip, exists := ingMap["ip"]; exists {
				if ipStr, ok := ip.(string); ok {
					ips = append(ips, ipStr)
				}
			} else if hostname, exists := ingMap["hostname"]; exists {
				if hostStr, ok := hostname.(string); ok {
					ips = append(ips, hostStr)
				}
			}
		}
	}

	if len(ips) > 0 {
		return strings.Join(ips, ",")
	}
	return "<pending>"
}

func extractPortsValue(obj *unstructured.Unstructured) string {
	ports, found, _ := unstructured.NestedSlice(obj.Object, "spec", "ports")
	if !found || len(ports) == 0 {
		return "<none>"
	}

	var portStrings []string
	for _, port := range ports {
		if portMap, ok := port.(map[string]interface{}); ok {
			if portNum, exists := portMap["port"]; exists {
				if p, ok := portNum.(int64); ok {
					portStrings = append(portStrings, fmt.Sprintf("%d", p))
				}
			}
		}
	}

	return strings.Join(portStrings, ",")
}

func extractUpToDateValue(obj *unstructured.Unstructured) string {
	updated, found, _ := unstructured.NestedInt64(obj.Object, "status", "updatedReplicas")
	if !found {
		return "0"
	}
	return fmt.Sprintf("%d", updated)
}

func extractAvailableValue(obj *unstructured.Unstructured) string {
	available, found, _ := unstructured.NestedInt64(obj.Object, "status", "availableReplicas")
	if !found {
		return "0"
	}
	return fmt.Sprintf("%d", available)
}

func extractDesiredValue(obj *unstructured.Unstructured) string {
	desired, found, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas")
	if !found {
		return "0"
	}
	return fmt.Sprintf("%d", desired)
}

func extractCurrentValue(obj *unstructured.Unstructured) string {
	current, found, _ := unstructured.NestedInt64(obj.Object, "status", "replicas")
	if !found {
		return "0"
	}
	return fmt.Sprintf("%d", current)
}

func extractCompletionsValue(obj *unstructured.Unstructured) string {
	succeeded, found1, _ := unstructured.NestedInt64(obj.Object, "status", "succeeded")
	completions, found2, _ := unstructured.NestedInt64(obj.Object, "spec", "completions")

	if !found1 {
		return "<none>"
	}

	if !found2 {
		return fmt.Sprintf("%d/1", succeeded)
	}

	return fmt.Sprintf("%d/%d", succeeded, completions)
}

func extractDurationValue(obj *unstructured.Unstructured) string {
	startTime, found, _ := unstructured.NestedString(obj.Object, "status", "startTime")
	completionTime, found2, _ := unstructured.NestedString(obj.Object, "status", "completionTime")

	if !found {
		return "<unknown>"
	}

	// Parse start time
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return "<unknown>"
	}

	var end time.Time
	if found2 {
		end, err = time.Parse(time.RFC3339, completionTime)
		if err != nil {
			return "<unknown>"
		}
	} else {
		end = time.Now()
	}

	duration := end.Sub(start)
	return duration.String()
}

func extractDataCountValue(obj *unstructured.Unstructured) string {
	data, found1, _ := unstructured.NestedMap(obj.Object, "data")
	binaryData, found2, _ := unstructured.NestedMap(obj.Object, "binaryData")

	count := 0
	if found1 {
		count += len(data)
	}
	if found2 {
		count += len(binaryData)
	}

	return fmt.Sprintf("%d", count)
}

func extractCapacityValue(obj *unstructured.Unstructured) string {
	capacity, found, _ := unstructured.NestedMap(obj.Object, "spec", "capacity")
	if !found {
		capacity, found, _ = unstructured.NestedMap(obj.Object, "status", "capacity")
	}

	if !found || len(capacity) == 0 {
		return "<unknown>"
	}

	// Return the first capacity value (usually storage)
	for _, v := range capacity {
		if val, ok := v.(string); ok {
			return val
		}
		if val, ok := v.(map[string]interface{}); ok {
			// Handle resource.Quantity format
			if s, exists := val["string"]; exists {
				if str, ok := s.(string); ok {
					return str
				}
			}
		}
	}

	return "<unknown>"
}

func extractAccessModesValue(obj *unstructured.Unstructured) string {
	modes, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "accessModes")
	if !found {
		modes, found, _ = unstructured.NestedStringSlice(obj.Object, "status", "accessModes")
	}

	if !found || len(modes) == 0 {
		return "<unknown>"
	}

	return strings.Join(modes, ",")
}

func extractClaimValue(obj *unstructured.Unstructured) string {
	claimRef, found, _ := unstructured.NestedMap(obj.Object, "spec", "claimRef")
	if !found {
		return "<none>"
	}

	name, found1, _ := unstructured.NestedString(claimRef, "name")
	namespace, found2, _ := unstructured.NestedString(claimRef, "namespace")

	if found1 && found2 {
		return fmt.Sprintf("%s/%s", namespace, name)
	} else if found1 {
		return name
	}

	return "<none>"
}

func extractStorageClassValue(obj *unstructured.Unstructured) string {
	sc, found, _ := unstructured.NestedString(obj.Object, "spec", "storageClassName")
	if !found {
		return "<none>"
	}
	return sc
}

func extractHostsValue(obj *unstructured.Unstructured) string {
	rules, found, _ := unstructured.NestedSlice(obj.Object, "spec", "rules")
	if !found || len(rules) == 0 {
		return "<none>"
	}

	var hosts []string
	for _, rule := range rules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			if host, exists := ruleMap["host"]; exists {
				if hostStr, ok := host.(string); ok && hostStr != "" {
					hosts = append(hosts, hostStr)
				}
			}
		}
	}

	if len(hosts) > 0 {
		return strings.Join(hosts, ",")
	}
	return "<none>"
}

func extractAddressValue(obj *unstructured.Unstructured) string {
	return extractExternalIPValue(obj) // Same logic as external IP
}

func extractIngressPortsValue(obj *unstructured.Unstructured) string {
	rules, found, _ := unstructured.NestedSlice(obj.Object, "spec", "rules")
	if !found || len(rules) == 0 {
		return "<none>"
	}

	portSet := make(map[string]struct{})
	for _, rule := range rules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			if http, exists := ruleMap["http"]; exists {
				if httpMap, ok := http.(map[string]interface{}); ok {
					if paths, exists := httpMap["paths"]; exists {
						if pathsSlice, ok := paths.([]interface{}); ok {
							for range pathsSlice {
								portSet["80"] = struct{}{}
							}
						}
					}
				}
			}
		}
	}

	// Check for TLS
	tls, found, _ := unstructured.NestedSlice(obj.Object, "spec", "tls")
	if found && len(tls) > 0 {
		portSet["443"] = struct{}{}
	}

	if len(portSet) > 0 {
		var ports []string
		for port := range portSet {
			ports = append(ports, port)
		}
		return strings.Join(ports, ",")
	}

	return "<none>"
}

func extractEndpointsValue(obj *unstructured.Unstructured) string {
	subsets, found, _ := unstructured.NestedSlice(obj.Object, "subsets")
	if !found || len(subsets) == 0 {
		return "<none>"
	}

	var endpoints []string
	for _, subset := range subsets {
		if subsetMap, ok := subset.(map[string]interface{}); ok {
			addresses, found1, _ := unstructured.NestedSlice(subsetMap, "addresses")
			ports, found2, _ := unstructured.NestedSlice(subsetMap, "ports")

			if found1 && found2 && len(addresses) > 0 && len(ports) > 0 {
				for _, addr := range addresses {
					if addrMap, ok := addr.(map[string]interface{}); ok {
						if ip, exists := addrMap["ip"]; exists {
							if ipStr, ok := ip.(string); ok {
								for _, port := range ports {
									if portMap, ok := port.(map[string]interface{}); ok {
										if portNum, exists := portMap["port"]; exists {
											if p, ok := portNum.(int64); ok {
												endpoints = append(endpoints, fmt.Sprintf("%s:%d", ipStr, p))
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if len(endpoints) > 0 {
		return strings.Join(endpoints, ",")
	}
	return "<none>"
}

func extractSecretsValue(obj *unstructured.Unstructured) string {
	secrets, found, _ := unstructured.NestedSlice(obj.Object, "secrets")
	if !found {
		return "0"
	}
	return fmt.Sprintf("%d", len(secrets))
}

func extractHardValue(obj *unstructured.Unstructured) string {
	hard, found, _ := unstructured.NestedMap(obj.Object, "status", "hard")
	return formatResourceLimits(hard, found)
}

func extractUsedValue(obj *unstructured.Unstructured) string {
	used, found, _ := unstructured.NestedMap(obj.Object, "status", "used")
	return formatResourceLimits(used, found)
}

func formatResourceLimits(limits map[string]interface{}, found bool) string {
	if !found || len(limits) == 0 {
		return "<none>"
	}

	var parts []string
	for key, value := range limits {
		if val, ok := value.(string); ok {
			parts = append(parts, fmt.Sprintf("%s:%s", key, val))
		} else if val, ok := value.(map[string]interface{}); ok {
			if s, exists := val["string"]; exists {
				if str, ok := s.(string); ok {
					parts = append(parts, fmt.Sprintf("%s:%s", key, str))
				}
			}
		}
	}

	if len(parts) > 0 {
		return strings.Join(parts, ",")
	}
	return "<none>"
}

func extractPodSelectorValue(obj *unstructured.Unstructured) string {
	selector, found, _ := unstructured.NestedMap(obj.Object, "spec", "podSelector", "matchLabels")
	if !found || len(selector) == 0 {
		return "<none>"
	}

	var labels []string
	for k, v := range selector {
		if val, ok := v.(string); ok {
			labels = append(labels, fmt.Sprintf("%s=%s", k, val))
		}
	}

	if len(labels) > 0 {
		return strings.Join(labels, ",")
	}
	return "<none>"
}

func extractPolicyTypesValue(obj *unstructured.Unstructured) string {
	types, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "policyTypes")
	if !found || len(types) == 0 {
		return "<none>"
	}
	return strings.Join(types, ",")
}

func extractLastSeenValue(obj *unstructured.Unstructured) string {
	lastTimestamp, found1, _ := unstructured.NestedString(obj.Object, "lastTimestamp")
	firstTimestamp, found2, _ := unstructured.NestedString(obj.Object, "firstTimestamp")

	var timestamp string
	if found1 {
		timestamp = lastTimestamp
	} else if found2 {
		timestamp = firstTimestamp
	} else {
		return "<unknown>"
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "<unknown>"
	}

	return fmt.Sprintf("%s ago", time.Since(t).Round(time.Second))
}

func extractObjectValue(obj *unstructured.Unstructured) string {
	kind, found1, _ := unstructured.NestedString(obj.Object, "involvedObject", "kind")
	name, found2, _ := unstructured.NestedString(obj.Object, "involvedObject", "name")

	if found1 && found2 {
		return fmt.Sprintf("%s/%s", kind, name)
	}
	return "<unknown>"
}

func extractNodeSelectorValue(obj *unstructured.Unstructured) string {
	selector, found, _ := unstructured.NestedMap(obj.Object, "spec", "template", "spec", "nodeSelector")
	if !found || len(selector) == 0 {
		return "<none>"
	}

	var selectors []string
	for k, v := range selector {
		if val, ok := v.(string); ok {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, val))
		}
	}

	if len(selectors) > 0 {
		return strings.Join(selectors, ",")
	}
	return "<none>"
}

func extractAllowVolumeExpansionValue(obj *unstructured.Unstructured) string {
	allow, found, _ := unstructured.NestedBool(obj.Object, "allowVolumeExpansion")
	if !found {
		return "false"
	}
	return fmt.Sprintf("%t", allow)
}

func extractScheduleValue(obj *unstructured.Unstructured) string {
	schedule, found, _ := unstructured.NestedString(obj.Object, "spec", "schedule")
	if !found {
		return "<none>"
	}
	return schedule
}

func extractSuspendValue(obj *unstructured.Unstructured) string {
	suspend, found, _ := unstructured.NestedBool(obj.Object, "spec", "suspend")
	if !found {
		return "False"
	}
	return fmt.Sprintf("%t", suspend)
}

func extractActiveValue(obj *unstructured.Unstructured) string {
	active, found, _ := unstructured.NestedSlice(obj.Object, "status", "active")
	if !found {
		return "0"
	}
	return fmt.Sprintf("%d", len(active))
}

func extractLastScheduleValue(obj *unstructured.Unstructured) string {
	lastSchedule, found, _ := unstructured.NestedString(obj.Object, "status", "lastScheduleTime")
	if !found {
		return "<none>"
	}

	t, err := time.Parse(time.RFC3339, lastSchedule)
	if err != nil {
		return "<none>"
	}

	return fmt.Sprintf("%s ago", time.Since(t).Round(time.Second))
}

func extractCreatedAtValue(obj *unstructured.Unstructured) string {
	created, found, _ := unstructured.NestedString(obj.Object, "metadata", "creationTimestamp")
	if !found {
		return "<unknown>"
	}

	t, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return "<unknown>"
	}

	return time.Since(t).String()
}
