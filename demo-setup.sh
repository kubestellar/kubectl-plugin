#!/bin/bash
# Demo setup script for kubectl-multi plugin using official KubeStellar demo environment

set -e

# Run the official KubeStellar demo environment
bash scripts/create-kubestellar-demo-env.sh --platform kind

echo "Demo environment setup complete."
echo "You can now run kubectl-multi get commands!"