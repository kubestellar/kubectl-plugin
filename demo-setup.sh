#!/bin/bash
# Demo setup script for kubectl-multi plugin
# This script assumes you have kind, kubectl, and kubectl-multi built and in your PATH

set -e

# Create two kind clusters for demo
kind create cluster --name cluster1
kind create cluster --name cluster2

# Set kubeconfig context names
kubectl config rename-context kind-cluster1 cluster1
kubectl config rename-context kind-cluster2 cluster2

# Deploy demo resources to both clusters
for ctx in cluster1 cluster2; do
  kubectl --context $ctx create namespace demo || true
  kubectl --context $ctx apply -f demo-deploy.yaml -n demo
  kubectl --context $ctx create deployment demo-deploy --image=nginx -n demo || true
  kubectl --context $ctx create daemonset demo-ds --image=nginx -n demo || true
  kubectl --context $ctx create statefulset demo-sts --image=nginx --replicas=1 --service-name=demo-sts -n demo || true
  kubectl --context $ctx create cronjob demo-cron --image=busybox --schedule="*/1 * * * *" -- /bin/sh -c 'date; echo Hello from the Kubernetes cluster' -n demo || true
  kubectl --context $ctx create serviceaccount demo-sa -n demo || true
  kubectl --context $ctx create ingress demo-ing --rule="demo.local/*=demo-deploy:80" -n demo || true
  kubectl --context $ctx create networkpolicy demo-np --pod-selector=app=demo-deploy --policy-types=Ingress -n demo || true
  # Add more demo resources as needed
  echo "Demo resources created in $ctx"
done

echo "Demo environment setup complete."
echo "You can now run kubectl-multi get commands!"
