#!/usr/bin/env bash
set -e

ZK_SERVICE_NAME=

find_leader() {
  for pod_name in $(oc get pod -o=custom-columns=NAME:.metadata.name | grep "zookeeper") ; do
    if oc exec $pod_name -- bin/zkServer.sh status | grep "Mode: leader" ; then
      eval "$1"="$pod_name"
      break
    fi
  done
}

run() {
  if [ -z "$ZK_SERVICE_NAME" ]; then
    echo "Error: zookeeper service name is not specified or empty!"
    exit 1
  fi
  
  local leader_node
  find_leader leader_node
  if [ -z "$leader_node" ]; then
    echo "Error: leader node is not found!"
    exit 1
  fi
  
  oc delete --now=true pod $leader_node
  echo "pod $leader_node is deleted"
  sleep 5
  
  local new_leader_node
  find_leader new_leader_node
  while [ -z "$new_leader_node" ]; do
    sleep 1
    find_leader new_leader_node
  done

  echo "previous leader node: $leader_node"
  echo "     new leader node: $new_leader_node"
}

run

