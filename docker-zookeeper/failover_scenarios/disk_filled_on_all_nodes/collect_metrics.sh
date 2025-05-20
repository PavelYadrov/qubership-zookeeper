#!/usr/bin/env bash
set -e

ZOOKEEPER_POD_NAME=

run() {
  if [ -z "$ZOOKEEPER_POD_NAME" ]; then
    echo "Error: ZooKeeper pod name is not specified or empty!"
    exit 1
  fi

  # KUBECONFIG works like --config
  export KUBECONFIG=$(dirname $0)"/admin.kubeconfig"
  while true; do
    oc exec ${ZOOKEEPER_POD_NAME} -- df -h /var/opt/zookeeper/data
    sleep 5
  done
}

run
