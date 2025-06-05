
### Demo

<img width="565" alt="image" src="https://github.com/user-attachments/assets/a0116533-c819-43f2-9bae-1274ccf70602" />


### TODO: clarify the requirement

> Create a kubectl plugin that is called "multi".
> Running "kubectl multi get nodes" returns the current nodes that are in my context in regular table output,
> followed by cluster nodes that are defined in the namespace ITS1 and are of type managedcluster.

- what is "cluster nodes that are defined in the namespace ITS1 and are of type managedcluster
# kubectl-multi plugin

`kubectl-multi` is a `kubectl` plugin that lets you view nodes from both your local Kubernetes cluster and remote managed clusters (e.g., from a KubeStellar universe), in a single unified tabular format.

---

## 🚀 Installation

Clone your fork and build the binary:

```bash
git clone https://github.com/kubestellar/kubectl-plugin.git

# Build the plugin
go build -o kubectl-multi

# Make it executable and move to your PATH
chmod +x kubectl-multi
sudo mv kubectl-multi /usr/local/bin/

# Now you can use it like a regular kubectl command:
kubectl multi get nodes

--remote-context string   # Remote context name to fetch managed clusters (default "its1")
--kubeconfig string       # Path to custom kubeconfig file (optional)

# Example:
kubectl multi get nodes --remote-context its1 --kubeconfig ~/.kube/config






