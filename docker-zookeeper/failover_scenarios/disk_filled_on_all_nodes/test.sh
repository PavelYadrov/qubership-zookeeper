#!/usr/bin/env bash
set -e

ZOOKEEPER_POD_NAME=
PROJECT_NAME=
OPENSHIFT_URL=
OPENSHIFT_USER=admin
OPENSHIFT_PASSWORD=admin

login_to_openshift() {
  echo "=> Login to OpenShift..."
  # KUBECONFIG works like --config
  export KUBECONFIG=$(dirname $0)"/admin.kubeconfig"
  oc login \
    --server= ${OPENSHIFT_URL} \
    --username=${OPENSHIFT_USER} \
    --password=${OPENSHIFT_PASSWORD} \
    --insecure-skip-tls-verify
  if [ $? -ne 0 ]; then
    echo >&2 "Login to openshift [$OPENSHIFT_URL] as user [$OPENSHIFT_USER] failed"
    exit 1
  fi
  oc project ${PROJECT_NAME}
}

run() {
  if [ -z "$ZOOKEEPER_POD_NAME" ]; then
    echo "Error: ZooKeeper pod name is not specified or empty!"
    exit 1
  fi
  if [ -z "$PROJECT_NAME" ]; then
    echo "Error: ZooKeeper project name is not specified or empty!"
    exit 1
  fi
  if [ -z "$OPENSHIFT_URL" ]; then
    echo >&2 "Error: OpenShift server URL is not specified or empty!"
    exit 1
  fi
  login_to_openshift

  echo "Create /test_failover_scenarios/disk_filled_on_all_nodes znode"
  oc exec ${ZOOKEEPER_POD_NAME} \
    -- ./bin/zkCli.sh create /test_failover_scenarios "test_failover_scenarios"
  oc exec ${ZOOKEEPER_POD_NAME} \
    -- ./bin/zkCli.sh create /test_failover_scenarios/disk_filled_on_all_nodes "disk_filled_on_all_nodes"

  echo "Copy test data"
  oc rsync ./test_files ${ZOOKEEPER_POD_NAME}:/opt/zookeeper --strategy='tar'
  oc exec ${ZOOKEEPER_POD_NAME} -- chmod +x ./test_files/*

  echo "Fill disk space"
  # You need to calculate size (in bytes) of "busy_space" file. Size must be greater than 90
  # percent of total size of all file stores.
  # It can be calculated by expression: (90 * disk.total) / 100
  # where disk.total is obtained from df -h /var/opt/zookeeper/data
  oc exec ${ZOOKEEPER_POD_NAME} -- dd if=/dev/zero of=/var/opt/zookeeper/data/busy_space bs=50M count=37

  local number=0
  while true; do
    echo "Create znode $number"
    oc exec ${ZOOKEEPER_POD_NAME} -- ./test_files/create_znode.sh ${number}
    echo "Znode $number is created!"
    number=$(( $number + 1 ))
  done
}

run
