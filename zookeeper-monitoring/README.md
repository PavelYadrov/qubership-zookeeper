# Telegraf Plugin: ZooKeeper

## Description

The ZooKeeper plugin collects variables outputted from the 'mntr' command.

```
echo mntr | nc localhost 2181

              zk_version  3.4.0
              zk_avg_latency  0
              zk_max_latency  0
              zk_min_latency  0
              zk_packets_received 70
              zk_packets_sent 69
              zk_outstanding_requests 0
              zk_server_state leader
              zk_znode_count   4
              zk_watch_count  0
              zk_ephemerals_count 0
              zk_approximate_data_size    27
              zk_followers    4                   - only exposed by the Leader
              zk_synced_followers 4               - only exposed by the Leader
              zk_pending_syncs    0               - only exposed by the Leader
              zk_open_file_descriptor_count 23    - only available on Unix platforms
              zk_max_file_descriptor_count 1024   - only available on Unix platforms
```

## Configuration

```
# Reads 'mntr' stats from one or many zookeeper servers
[[inputs.zookeeper]]
  ## An array of address to gather stats about. Specify an ip or hostname
  ## with port. ie localhost:2181, 10.0.0.1:2181, etc.

  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 2181 is used
  servers = [":2181"]
```

## Grafana
### Dashboard exporting
```bash
curl -XGET -k -u admin:admin http://localhost:3000/api/dashboards/db/zookeeper-monitoring \
  | jq . > dashboard/zookeeper-dashboard.json
```

  Where:
   
   * `admin:admin` grafana user login and password
   * `http://localhost:3000` grafana url
   * `zookeeper-monitoring` dashboard name
 
### Dashboard importing
Dashboard can be imported using the following command:

```bash
curl -XPOST \
  -u admin:admin \
  --data @./dashboard/zookeeper-dashboard.json \
  -H 'Content-Type: application/json'  \
  -k \
   http://localhost:3000/api/dashboards/db
```
  Where:
   
   * `admin:admin` grafana user login and password
   * `http://localhost:3000` grafana url
   
## Zabbix
Zabbix template for monitoring zookeeper cluster state consists of items and triggers for monitoring CPU usage, memory usage and state of the cluster.

Triggers use following macroses to control thresholds for CPU and memory usage: {$ZOOKEEPER_CPU_THRESHOLD}, {$ZOOKEEPER_MEMORY_THRESHOLD}. By default they are set to 0.95 so triggers will activate when 95% of available pod resources are used.


### Importing Template
Template can be imported in Zabbix UI from templates page (Configuration -> Templates) by using Import button.

### DR Mode
If you have Zookeeper which is deployed in DR mode you need to create two hosts: 
for left and for right side and to specify the side as value (`left`, `right`) for the macros `{$DR_SIDE}`.
If you have Zookeeper without DR just leave this macros empty.

## InfluxDB Measurement:

### Telegraf tags and field
```
M zookeeper
  T host
  T port
  T state
  
  F approximate_data_size        integer
  F avg_latency                  integer
  F ephemerals_count             integer
  F max_file_descriptor_count    integer
  F max_latency                  integer
  F min_latency                  integer
  F num_alive_connections        integer
  F open_file_descriptor_count   integer
  F outstanding_requests         integer
  F packets_received             integer
  F packets_sent                 integer
  F version                      string
  F watch_count                  integer
  F znode_count                  integer
  F status_code                  integer
  F alive_nodes                  integer
  F total_nodes                  integer
```


### Custom tags and field
Custom fields collected by `exec-scripts/health_metric.py`:

```
M zookeeper
  F status_code                  integer          0 - if all nodes alive; 5 - if some nodes failed; 10 - if all nodes failed
  F alive_nodes                  integer          number of alive nodes
  F total_nodes                  integer          total number of nodes
```
