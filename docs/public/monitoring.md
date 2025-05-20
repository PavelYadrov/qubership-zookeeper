This document describes the process of monitoring dashboards, their metrics and Prometheus alerts.

The dashboards provide the following parameters to configure at the top of the dashboard:

* Interval time for metric display
* Node name

For all graph panels, the mean metric value is used in the given interval. For all singlestat panels, the last metric value is used.

# ZooKeeper Monitoring

## Dashboard

![Dashboard](/docs/public/images/zookeeper-monitoring_dashboard.png)

## Metrics

[Metrics Overview](/docs/public/zookeeper-dashboard.md)

# Table of Metrics

This table provides the full list of Prometheus metrics being collected by ZooKeeper Monitoring.

| Metric name                                         | Description                                                                                                                                                   |             Source             | ZooKeeper Service |
|-----------------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------------------:|:-----------------:|
| zookeeper_status_code                               | Status of the ZooKeeper cluster.                                                                                                                              |      Telegraf exec plugin      |     Supported     |
| zookeeper_alive_nodes                               | Current count of ZooKeeper servers in quorum.                                                                                                                 |      Telegraf exec plugin      |     Supported     |
| zookeeper_QuorumSize                                | Count of ZooKeeper servers in quorum. Contain information about leader server in labels.                                                                      | ZooKeeper Prometheus Exporter  |     Supported     |
| kube_pod_status_ready                               | Current count of ready pods.                                                                                                                                  | Kubernetes Prometheus exporter |     Supported     |
| zookeeper_znode_count                               | The number of unique z-nodes on a ZooKeeper server has in its data tree.                                                                                      |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_ephemerals_count                          | The number of unique ephemeral z-nodes on a ZooKeeper server.                                                                                                 |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_approximate_data_size                     | The size in bytes of the data tree for a ZooKeeper server.                                                                                                    |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_avg_latency                               | How long on average it takes for this ZooKeeper server to process a request in milliseconds. This is measured since the last restart of the ZooKeeper server. |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_min_latency                               | The minimum time it took this ZooKeeper server to process a request in milliseconds. This is measured since the last restart of this ZooKeeper server.        |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_max_latency                               | The maximum time it took this ZooKeeper server to process a request in milliseconds. This is measured since the last restart of the ZooKeeper server.         |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_num_alive_connections                     | The number of active clients connected to a ZooKeeper server.                                                                                                 |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_packets_received                          | How many ZooKeeper packets have been received by a ZooKeeper server.                                                                                          |   telegraf ZooKeeper plugin    |     Supported     |
| zookeeper_packets_sent                              | How many ZooKeeper packets have been sent from a ZooKeeper server.                                                                                            |   telegraf ZooKeeper plugin    |     Supported     |
| jvm_gc_collection_seconds_count                     | The number of garbage collections grouped by collection type.                                                                                                 | ZooKeeper Prometheus Exporter  |     Supported     |
| jvm_gc_collection_seconds_sum                       | The time spent on garbage collections grouped by collection type.                                                                                             | ZooKeeper Prometheus Exporter  |     Supported     |
| jvm_memory_bytes_used                               | This widget displays the general data on heap\non-heap memory usage.                                                                                          | ZooKeeper Prometheus Exporter  |     Supported     |
| jvm_memory_pool_bytes_used                          | JVM heap memory used by pools.                                                                                                                                | ZooKeeper Prometheus Exporter  |     Supported     |
| jvm_memory_pool_bytes_max                           | JVM heap memory limits by pools.                                                                                                                              | ZooKeeper Prometheus Exporter  |     Supported     |
| jvm_threads_total                                   | This widget displays the data about the thread count.                                                                                                         | ZooKeeper Prometheus Exporter  |     Supported     |
| kubelet_volume_stats_used_bytes                     | Number of used bytes in the volume.                                                                                                                           | Kubernetes Prometheus Exporter |     Supported     |
| kubelet_volume_stats_capacity_bytes                 | Capacity in bytes of the volume.                                                                                                                              | Kubernetes Prometheus Exporter |     Supported     |
| zookeeper_backup_metric_status                      | Status of the ZooKeeper Backup daemon.                                                                                                                        |      Telegraf exec plugin      |     Supported     |
| zookeeper_backup_metric_last_backup_status          | Shows the state of the last backup.                                                                                                                           |      Telegraf exec plugin      |     Supported     |
| zookeeper_backup_metric_successful_backups_count    | Shows the amount of successful backups.                                                                                                                       |      Telegraf exec plugin      |     Supported     |
| zookeeper_backup_metric_backups_count               | Shows the amount of available backups.                                                                                                                        |      Telegraf exec plugin      |     Supported     |
| zookeeper_backup_metric_last_backup_time            | Shows the period of time when the last backup process was ended.                                                                                              |      Telegraf exec plugin      |     Supported     |
| zookeeper_backup_metric_last_successful_backup_time | Shows the period of time when the last successful backup process was ended.                                                                                   |      Telegraf exec plugin      |     Supported     |
| zookeeper_backup_metric_storage_type                | Shows the backup storage type.                                                                                                                                |      Telegraf exec plugin      |     Supported     |
| service:tls_status:info                             | Shows the status of TLS for service.                                                                                                                          |         Static metric          |     Supported     |

## Monitoring Alarms Description

This section describes Prometheus monitoring alarms.

<!-- markdownlint-disable line-length -->
| Name                                   | Summary                                                           | For | Severity | Expression Example                                                                                                                                                                                                                                                                | Description                                       | Troubleshooting Link                                                               |
|----------------------------------------|-------------------------------------------------------------------|-----|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------|------------------------------------------------------------------------------------|
| ZooKeeper_Memory_Usage_Alarm           | Some of ZooKeeper Service pod uses memory higher then 95 percents | 3m  | high     | `max(container_memory_working_set_bytes{namespace="zookeeper\-service",container!\~"POD&#124;",pod=\~".\*zookeeper\-[0\-9].\*"}) / max(kube_pod_container_resource_limits_memory_bytes{exported_namespace="zookeeper\-service",exported_pod=\~".\*zookeeper\-[0\-9].\*"}) \>0.95` | ZooKeeper memory usage is higher than 95 percents | [ZooKeeper_Memory_Usage_Alarm](/docs/public/troubleshooting.md#alerts-3)           |
| ZooKeeper_Last_Backup_Has_Failed_Alarm | ZooKeeper Last Backup Has Failed                                  | 3m  | low      | `zookeeper_backup_metric_last_backup_status{host=\~"^.\*",project_name="zookeeper\-service"} == 4`                                                                                                                                                                                | ZooKeeper Last Backup Has Failed                  | [ZooKeeper_Last_Backup_Has_Failed_Alarm](/docs/public/troubleshooting.md#alerts-5) |
| ZooKeeper_Is_Down_Alarm                | All of ZooKeeper Service pods are down                            | 3m  | disaster | `zookeeper_status_code{host=\~"^.\*",project_name="zookeeper\-service"} == 10`                                                                                                                                                                                                    | ZooKeeper is Down.                                | [ZooKeeper_Is_Down_Alarm](/docs/public/troubleshooting.md#alerts-1)                |
| ZooKeeper_Is_Degraded_Alarm            | Some of ZooKeeper Service pods are down                           | 3m  | high     | `zookeeper_status_code{host=\~"^.\*",project_name="zookeeper\-service"} == 5`                                                                                                                                                                                                     | ZooKeeper is Degraded.                            | [ZooKeeper_Is_Degraded_Alarm](/docs/public/troubleshooting.md#alerts)              |
| ZooKeeper_CPU_Load_Alarm               | Some of ZooKeeper Service pod loads CPU higher then 95 percents   | 3m  | high     | `max(rate(container_cpu_usage_seconds_total{namespace="zookeeper\-service", pod=\~".\*zookeeper\-[0\-9].\*"}[1m])) / max(kube_pod_container_resource_limits_cpu_cores{exported_namespace="zookeeper\-service", exported_pod=\~".\*zookeeper\-[0\-9].\*"}) \> 0.95`                | ZooKeeper CPU load is higher than 95 percents     | [ZooKeeper_CPU_Load_Alarm](/docs/public/troubleshooting.md#alerts-2)               |
<!-- markdownlint-enable line-length -->
