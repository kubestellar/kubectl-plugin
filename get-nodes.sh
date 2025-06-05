#!/usr/bin/env bash
CURR_CTX=$(kubectl config current-context)

FMT="%-6s | %-20s | %-15s | %-28s\n"

print_row() { printf "$FMT" "$@"; }

{
  print_row "CTX" "RESOURCE" "TYPE" "AGE"
  printf -- "%0.s-" {1..78}; echo        

  # local cluster
  print_row "$CURR_CTX" "$CURR_CTX" "CLUSTER" "-"

  # remote nodes
  kubectl get nodes -o custom-columns=NAME:.metadata.name,AGE:.metadata.creationTimestamp \
    --no-headers 2>/dev/null |
  while read -r node age; do
      print_row "" "  $node" "NODE" "$age"
  done

  # remote ManagedCluster
  kubectl --context "$REMOTE_CTX" get managedclusters \
    -o custom-columns=NAME:.metadata.name,AGE:.metadata.creationTimestamp \
    --no-headers 2>/dev/null |
  while read -r mc age; do
      print_row "$REMOTE_CTX" "$mc" "REMOTE-CLUSTER" "$age"
  done
} | column -t -s'|'
