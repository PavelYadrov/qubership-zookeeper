#!/usr/bin/env bash
set -e

create_znode() {
  echo "Input: $1"
  local data=$( tr -dc A-Za-z0-9 </dev/urandom | head -c 100000 )
  source ./bin/zkCli.sh create /test_failover_scenarios/disk_filled_on_all_nodes/$1 "$data"
}

create_znode "$@"