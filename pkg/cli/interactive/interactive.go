package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// InteractiveCLI handles the interactive mode
type InteractiveCLI struct {
	scanner *bufio.Scanner
	cyan    *color.Color
	bold    *color.Color
	dim     *color.Color
	green   *color.Color
	yellow  *color.Color
	red     *color.Color
}

// NewInteractiveCLI creates a new interactive CLI instance
func NewInteractiveCLI() *InteractiveCLI {
	return &InteractiveCLI{
		scanner: bufio.NewScanner(os.Stdin),
		cyan:    color.New(color.FgCyan),
		bold:    color.New(color.Bold, color.FgCyan),
		dim:     color.New(color.Faint),
		green:   color.New(color.FgGreen),
		yellow:  color.New(color.FgYellow),
		red:     color.New(color.FgRed),
	}
}

// Run starts the interactive CLI
func (cli *InteractiveCLI) Run() {
	cli.displayBanner()
	cli.displayWelcome()
	cli.runCommandLoop()
}

// displayBanner shows the KubeStellar ASCII art banner
func (cli *InteractiveCLI) displayBanner() {
	fmt.Println()
	cli.cyan.Println("╭─────────────────────────────────────────────────────────────────────────────────────────────╮")
	cli.cyan.Print("│")
	fmt.Print("                                                                                             ")
	cli.cyan.Println("│")
	cli.cyan.Print("│  ")
	cli.bold.Print("██╗  ██╗██╗   ██╗██████╗ ███████╗███████╗████████╗███████╗███████╗███████╗██╗      █████╗ ██████╗ ")
	cli.cyan.Println("│")
	cli.cyan.Print("│  ")
	cli.bold.Print("██║ ██╔╝██║   ██║██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔════╝██╔════╝██╔════╝██║     ██╔══██╗██╔══██╗")
	cli.cyan.Println("│")
	cli.cyan.Print("│  ")
	cli.bold.Print("█████╔╝ ██║   ██║██████╔╝█████╗  ███████╗   ██║   █████╗  ███████╗███████╗██║     ███████║██████╔╝")
	cli.cyan.Println("│")
	cli.cyan.Print("│  ")
	cli.bold.Print("██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══╝  ╚════██║╚════██║██║     ██╔══██║██╔══██╗")
	cli.cyan.Println("│")
	cli.cyan.Print("│  ")
	cli.bold.Print("██║  ██╗╚██████╔╝██████╔╝███████╗███████║   ██║   ███████╗███████║███████║███████╗██║  ██║██║  ██║")
	cli.cyan.Println("│")
	cli.cyan.Print("│  ")
	cli.bold.Print("╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝")
	cli.cyan.Println("│")
	cli.cyan.Print("│")
	fmt.Print("                                                                                             ")
	cli.cyan.Println("│")
	cli.cyan.Print("│                       ")
	cli.dim.Print("🌟 Multi-Cluster Kubernetes Management Agent 🌟")
	cli.cyan.Println("                       │")
	cli.cyan.Println("╰─────────────────────────────────────────────────────────────────────────────────────────────╯")
	fmt.Println()
}

// displayWelcome shows the welcome message and available commands
func (cli *InteractiveCLI) displayWelcome() {
	cli.green.Println("Welcome to KubeStellar Interactive CLI!")
	fmt.Println()
	cli.yellow.Println("Available commands:")
	fmt.Println("  clusters          - Manage clusters in ITS")
	fmt.Println("  bindingpolicy     - Manage binding policies")
	fmt.Println("  get               - Get resources from clusters")
	fmt.Println("  apply             - Apply configurations to clusters")
	fmt.Println("  helm              - Deploy Helm charts")
	fmt.Println("  config            - Manage configurations")
	fmt.Println("  help              - Show help for commands")
	fmt.Println("  exit              - Exit interactive mode")
	fmt.Println()
}

// runCommandLoop runs the main command loop
func (cli *InteractiveCLI) runCommandLoop() {
	for {
		cli.green.Print("kubestellar> ")
		
		if !cli.scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(cli.scanner.Text())
		if input == "" {
			continue
		}
		
		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]
		
		switch command {
		case "exit", "quit", "q":
			cli.yellow.Println("Goodbye! 👋")
			return
			
		case "help", "h", "?":
			cli.showHelp(args)
			
		case "clusters", "cluster", "c":
			cli.handleClustersCommand(args)
			
		case "bindingpolicy", "bp", "policy", "binding":
			cli.handleBindingPolicyCommand(args)
			
		case "get", "g":
			cli.handleGetCommand(args)
			
		case "apply", "app":
			cli.handleApplyCommand(args)
			
		case "delete", "del", "rm":
			cli.handleDeleteCommand(args)
			
		case "helm", "hl":
			cli.handleHelmCommand(args)
			
		case "config", "cfg":
			cli.handleConfigCommand(args)
			
		case "clear", "cls":
			fmt.Print("\033[2J\033[H") // Clear screen
			cli.displayBanner()
			
		default:
			cli.red.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'help' for available commands or try short forms like 'c' for clusters, 'bp' for bindingpolicy, 'g' for get")
		}
		fmt.Println()
	}
}

// showHelp displays help information
func (cli *InteractiveCLI) showHelp(args []string) {
	if len(args) == 0 {
		cli.displayWelcome()
		return
	}
	
	switch args[0] {
	case "clusters", "cluster", "c":
		cli.yellow.Println("Cluster Management Commands (short form: c):")
		fmt.Println("  clusters list                      - List all clusters in ITS")
		fmt.Println("  clusters add <name> [options]      - Register a new cluster")
		fmt.Println("  clusters remove <name>             - Remove a cluster from ITS")
		fmt.Println("  clusters info <name>               - Show cluster details")
		fmt.Println("  clusters label <name> <labels>     - Update cluster labels")
		fmt.Println("  c ls                               - Short form for list")
		fmt.Println("  c add prod-cluster                 - Short form for add")
		
	case "bindingpolicy", "bp", "policy", "binding":
		cli.yellow.Println("BindingPolicy Commands (short forms: bp, policy, binding):")
		fmt.Println("  bindingpolicy list                           - List all binding policies")
		fmt.Println("  bindingpolicy create <name> [options]        - Create a new binding policy")
		fmt.Println("  bindingpolicy delete <name>                  - Delete a binding policy")
		fmt.Println("  bindingpolicy add-clusters <policy> <clusters...>    - Add clusters to policy")
		fmt.Println("  bindingpolicy remove-clusters <policy> <clusters...> - Remove clusters from policy")
		fmt.Println("  bindingpolicy update-labels <policy> [options]       - Update label selectors")
		fmt.Println("  bp create web-policy --cluster-labels env=prod       - Create with cluster labels")
		fmt.Println("  bp update-labels app-policy --workload-labels app=nginx - Update workload labels")
		
	case "helm", "h":
		cli.yellow.Println("Helm Commands (short form: h):")
		fmt.Println("  helm install <release> <chart>       - Install a Helm chart")
		fmt.Println("  helm upgrade <release> <chart>       - Upgrade a Helm release")
		fmt.Println("  helm list                            - List Helm releases")
		fmt.Println("  helm uninstall <release>             - Uninstall a Helm release")
		fmt.Println("  helm values <release>                - Show values for a release")
		fmt.Println("  helm rollback <release> <revision>   - Rollback to a previous revision")
		fmt.Println("  h install nginx nginx/nginx          - Short form for install")
		fmt.Println("  h ls                                 - Short form for list")
		
	case "get", "g":
		cli.yellow.Println("Get Commands (short form: g):")
		fmt.Println("  get pods                             - Get pods from all clusters")
		fmt.Println("  get nodes                            - Get nodes from all clusters")
		fmt.Println("  get deployments -n production        - Get deployments in namespace")
		fmt.Println("  g pods -l app=nginx                  - Short form with label selector")
		fmt.Println("  g all -A                             - Short form for all resources in all namespaces")
		
	case "apply", "app":
		cli.yellow.Println("Apply Commands (short form: app):")
		fmt.Println("  apply -f deployment.yaml             - Apply configuration to all clusters")
		fmt.Println("  apply -f dir/                        - Apply all files in directory")
		fmt.Println("  app -f manifest.yaml                 - Short form for apply")
		
	case "delete", "del", "rm":
		cli.yellow.Println("Delete Commands (short forms: del, rm):")
		fmt.Println("  delete pod nginx-pod                 - Delete pod from all clusters")
		fmt.Println("  delete pods -l app=nginx             - Delete pods with label selector")
		fmt.Println("  del deployment webapp                - Short form for delete")
		fmt.Println("  rm service old-service               - Short form for delete")
		
	case "config", "cfg":
		cli.yellow.Println("Configuration Commands (short form: cfg):")
		fmt.Println("  config set <key> <value>             - Set a configuration value")
		fmt.Println("  config get <key>                     - Get a configuration value")
		fmt.Println("  config list                          - List all configurations")
		fmt.Println("  config apply <file>                  - Apply configuration from YAML file")
		fmt.Println("  config export <file>                 - Export configuration to YAML file")
		fmt.Println("  cfg set its-context my-its           - Short form for config set")
		
	default:
		cli.red.Printf("No help available for: %s\n", args[0])
		fmt.Println("Available command groups: clusters (c), bindingpolicy (bp), helm (h), get (g), apply (app), delete (del), config (cfg)")
	}
}

// handleClustersCommand handles cluster management commands
func (cli *InteractiveCLI) handleClustersCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a subcommand. Type 'help clusters' for available commands.")
		return
	}
	
	switch args[0] {
	case "list":
		cli.green.Println("Listing clusters in ITS...")
		// TODO: Call ListClustersInITS from operations
		fmt.Println("NAME       TYPE    STATUS    ENDPOINT")
		fmt.Println("cluster1   wec     Ready     https://cluster1.example.com")
		fmt.Println("cluster2   wec     Ready     https://cluster2.example.com")
		
	case "add":
		if len(args) < 3 {
			cli.red.Println("Usage: clusters add <name> <endpoint>")
			return
		}
		cli.green.Printf("Adding cluster %s with endpoint %s...\n", args[1], args[2])
		// TODO: Call RegisterClusterWithITS from operations
		
	case "remove":
		if len(args) < 2 {
			cli.red.Println("Usage: clusters remove <name>")
			return
		}
		cli.green.Printf("Removing cluster %s...\n", args[1])
		// TODO: Call RemoveClusterFromITS from operations
		
	case "info":
		if len(args) < 2 {
			cli.red.Println("Usage: clusters info <name>")
			return
		}
		cli.green.Printf("Getting info for cluster %s...\n", args[1])
		// TODO: Call GetClusterDetails from operations
		
	default:
		cli.red.Printf("Unknown clusters subcommand: %s\n", args[0])
	}
}

// handleBindingPolicyCommand handles binding policy commands
func (cli *InteractiveCLI) handleBindingPolicyCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a subcommand. Type 'help bindingpolicy' for available commands.")
		return
	}
	
	switch args[0] {
	case "list":
		cli.green.Println("Listing binding policies...")
		// TODO: Call ListBindingPolicies from operations
		fmt.Println("NAMESPACE    NAME           CLUSTERS")
		fmt.Println("default      app-policy     all")
		fmt.Println("production   prod-policy    2 selector(s)")
		
	case "create":
		if len(args) < 2 {
			cli.red.Println("Usage: bindingpolicy create <name>")
			return
		}
		cli.green.Printf("Creating binding policy %s...\n", args[1])
		// TODO: Call CreateBindingPolicy from operations
		
	case "delete":
		if len(args) < 2 {
			cli.red.Println("Usage: bindingpolicy delete <name>")
			return
		}
		cli.green.Printf("Deleting binding policy %s...\n", args[1])
		// TODO: Call DeleteBindingPolicy from operations
		
	case "add-clusters":
		if len(args) < 3 {
			cli.red.Println("Usage: bindingpolicy add-clusters <policy> <clusters...>")
			return
		}
		cli.green.Printf("Adding clusters to policy %s: %v\n", args[1], args[2:])
		// TODO: Call UpdateBindingPolicyWithClusters from operations
		
	default:
		cli.red.Printf("Unknown bindingpolicy subcommand: %s\n", args[0])
	}
}

// handleGetCommand handles get commands
func (cli *InteractiveCLI) handleGetCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a resource type. Example: get pods")
		return
	}
	
	cli.green.Printf("Getting %s from all clusters...\n", args[0])
	// TODO: Call ExecuteGet from operations
	fmt.Printf("CLUSTER    NAME         STATUS    AGE\n")
	fmt.Printf("cluster1   nginx-pod    Running   2d\n")
	fmt.Printf("cluster2   nginx-pod    Running   2d\n")
}

// handleApplyCommand handles apply commands
func (cli *InteractiveCLI) handleApplyCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a file. Example: apply deployment.yaml")
		return
	}
	
	cli.green.Printf("Applying %s to all clusters...\n", args[0])
	// TODO: Call ExecuteApply from operations
	fmt.Println("deployment.apps/nginx created in cluster1")
	fmt.Println("deployment.apps/nginx created in cluster2")
}

// handleHelmCommand handles Helm commands
func (cli *InteractiveCLI) handleHelmCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a subcommand. Type 'help helm' for available commands.")
		return
	}
	
	switch args[0] {
	case "install":
		if len(args) < 3 {
			cli.red.Println("Usage: helm install <release> <chart>")
			return
		}
		cli.green.Printf("Installing Helm chart %s as %s...\n", args[2], args[1])
		// TODO: Call HelmInstall from operations
		
	case "list":
		cli.green.Println("Listing Helm releases...")
		// TODO: Call HelmList from operations
		fmt.Println("NAME       NAMESPACE    REVISION    STATUS      CHART")
		fmt.Println("nginx      default      1           deployed    nginx-1.0.0")
		
	case "uninstall":
		if len(args) < 2 {
			cli.red.Println("Usage: helm uninstall <release>")
			return
		}
		cli.green.Printf("Uninstalling Helm release %s...\n", args[1])
		// TODO: Call HelmUninstall from operations
		
	default:
		cli.red.Printf("Unknown helm subcommand: %s\n", args[0])
	}
}

// handleConfigCommand handles configuration commands
func (cli *InteractiveCLI) handleConfigCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a subcommand. Type 'help config' for available commands.")
		return
	}
	
	switch args[0] {
	case "set":
		if len(args) < 3 {
			cli.red.Println("Usage: config set <key> <value>")
			return
		}
		cli.green.Printf("Setting %s = %s\n", args[1], args[2])
		// TODO: Implement config management
		
	case "get":
		if len(args) < 2 {
			cli.red.Println("Usage: config get <key>")
			return
		}
		cli.green.Printf("Getting config value for %s\n", args[1])
		// TODO: Implement config management
		
	case "list":
		cli.green.Println("Configuration values:")
		fmt.Println("its-context: its1")
		fmt.Println("wds-context: wds1")
		fmt.Println("default-namespace: default")
		
	case "apply":
		if len(args) < 2 {
			cli.red.Println("Usage: config apply <file>")
			return
		}
		cli.green.Printf("Applying configuration from %s\n", args[1])
		// TODO: Implement config file loading
		
	default:
		cli.red.Printf("Unknown config subcommand: %s\n", args[0])
	}
}

// handleDeleteCommand handles delete commands
func (cli *InteractiveCLI) handleDeleteCommand(args []string) {
	if len(args) == 0 {
		cli.red.Println("Please specify a resource type to delete. Examples:")
		fmt.Println("  delete pod nginx-pod")
		fmt.Println("  delete deployments -l app=nginx")
		fmt.Println("  del service webapp")
		return
	}
	
	resourceType := args[0]
	resourceArgs := args[1:]
	
	cli.green.Printf("Deleting %s across all clusters", resourceType)
	if len(resourceArgs) > 0 {
		fmt.Printf(" with args: %s", strings.Join(resourceArgs, " "))
	}
	fmt.Println()
	
	// TODO: Implement actual delete operation using core operations
	fmt.Printf("CLUSTER    RESOURCE     STATUS\n")
	fmt.Printf("cluster1   %s          Deleted\n", resourceType)
	fmt.Printf("cluster2   %s          Deleted\n", resourceType)
	fmt.Printf("cluster3   %s          Deleted\n", resourceType)
}