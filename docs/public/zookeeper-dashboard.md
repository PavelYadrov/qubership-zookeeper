# ZooKeeper Monitoring

ZooKeeper Prometheus Monitoring

## Tags

* `Prometheus`
* `ZooKeeper`

## Panels

### Cluster Overview

![Cluster Overview](/docs/public/images/zookeeper-monitoring_cluster_overview.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Cluster Status | Status of ZooKeeper cluster | Default:<br/>Mode: absolute<br/>Level 1: 1<br/>Level 2: 9<br/><br/> |  |
| Quorum Size | Current count of ZooKeeper servers in quorum. |  |  |
| Ready Pods | Current count of ready pods. | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| Leader Server | Current leader server. | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| CPU Usage | Maximum current CPU usage (in percent) among all ZooKeeper servers. |  |  |
| Heap Usage | Maximum current JVM heap usage (in percent) among all ZooKeeper servers |  |  |
| Cluster Status Transitions | Transitions of ZooKeeper cluster statuses |  |  |
| Pod Readiness Probe Transitions | Transitions of readiness probes for each ZooKeeper pod. |  |  |
| Zookeeper Version | The metric indicates the specific version of the Zookeeper currently running. | | |
<!-- markdownlint-enable line-length -->

### System Overview

![System Overview](/docs/public/images/zookeeper-monitoring_system_overview.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Pod Memory Usage | Memory usage by pod |  |  |
| Pod CPU Usage | CPU usage by pod |  |  |
<!-- markdownlint-enable line-length -->

### Data Overview

![Data Overview](/docs/public/images/zookeeper-monitoring_data_overview.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Number of ZNodes | The number of unique z-nodes on a ZooKeeper server has in its data tree |  |  |
| Number of Ephemeral ZNodes | The number of unique ephemeral z-nodes on a ZooKeeper server |  |  |
| Size of Data | The size in bytes of the data tree for a ZooKeeper server |  |  |
| Disk usage in percent | The data usage size in percent for a ZooKeeper server |  |  |
| Disk usage | The data usage size in bytes for a ZooKeeper server |  |  |
<!-- markdownlint-enable line-length -->

### Latency

![Latency](/docs/public/images/zookeeper-monitoring_latency.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Average Request Latency | How long on average it takes for this ZooKeeper server to process a request in milliseconds. This is measured since the last restart of the ZooKeeper server |  |  |
| Minimum Request Latency | The minimum time it took this ZooKeeper server to process a request in milliseconds. This is measured since the last restart of this ZooKeeper server |  |  |
| Maximum Request Latency | The maximum time it took this ZooKeeper server to process a request in milliseconds. This is measured since the last restart of the ZooKeeper server |  |  |
<!-- markdownlint-enable line-length -->

### Connections

![Connections](/docs/public/images/zookeeper-monitoring_connections.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Number of Active Clients | The number of active clients connected to a ZooKeeper server |  |  |
| ZooKeeper Packets Received | How many ZooKeeper packets have been received by a ZooKeeper server |  |  |
| ZooKeeper Packets Sent | How many ZooKeeper packets have been sent from a ZooKeeper server |  |  |
<!-- markdownlint-enable line-length -->

### GC Metrics

![GC Metrics](/docs/public/images/zookeeper-monitoring_gc_metrics.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| GC Collection Count | This widget displays the number of garbage collections grouped by collection type. The statistics take into account two types of garbage collections: Copy - the garbage collector moves objects with links from eden space to survivor space and from survivor space to tenured space, <br/>MarkSweepCompact - the garbage collector works in tenured space, live objects are moved to the beginning, so garbage remains at the end of memory. |  |  |
| GC Collection Time | This widget displays the time spent on garbage collections grouped by collection type. <br/>The statistics take into account two types of garbage collections: Copy - the garbage collector moves objects with links from eden space to survivor space and from survivor space to tenured space, <br/>MarkSweepCompact - the garbage collector works in tenured space, live objects are moved to the beginning, so garbage remains at the end of memory. |  |  |
<!-- markdownlint-enable line-length -->

### Heap

![Heap](/docs/public/images/zookeeper-monitoring_heap.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Heap Memory Usage | This widget displays the general data on heap memory usage. |  |  |
| Heap Memory Usage In Percent | This widget displays the general data on heap memory usage in percent (%). |  |  |
| JVM Memory Pool | JVM Memory Pool consumption and limits. | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
<!-- markdownlint-enable line-length -->

### Non Heap

![Non Heap](/docs/public/images/zookeeper-monitoring_non_heap.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Non-Heap Memory Usage | This widget displays the general data on non heap memory usage. |  |  |
<!-- markdownlint-enable line-length -->

### Threads

![Threads](/docs/public/images/zookeeper-monitoring_thread_count.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Thread Count | This widget displays the data about the thread count. |  |  |
<!-- markdownlint-enable line-length -->

### Backup

![Backup](/docs/public/images/zookeeper-monitoring_backup.png)

<!-- markdownlint-disable line-length -->
| Name | Description | Thresholds | Repeat |
| ---- | ----------- | ---------- | ------ |
| Backup Daemon Status | Shows the current activity status of the backup daemon. The activity status can be one of the following:<br/>* UP - There is at least one successful backup, and the latest backup is successful<br/>* WARNING - There are no successful backups, or the last backup failed<br/>* NOT ACTIVE - The backup is not satisfied<br/>* MISSING - The host of ZooKeeper Backup Daemon is not specified | Default:<br/>Mode: absolute<br/>Level 1: 3<br/><br/> |  |
| Last Backup Status | Shows the state of the last backup. The backup status can be one of the following:<br/> <br/>* SUCCESS - The latest backup is successful<br/>* FAILED - The latest backup failed<br/>* IN QUEUE - The latest backup is in queue<br/>* IN PROGRESS - The latest backup is in progress | Default:<br/>Mode: absolute<br/>Level 1: 1<br/>Level 2: 4<br/><br/> |  |
| Successful Backups Count | Shows the amount of successful backups | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| Backups Count | Shows the amount of available backups | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| Backup Daemon Status Transitions | Transitions of ZooKeeper Backup Daemon statuses |  |  |
| Time of Last Backup | Shows the period of time when the last backup process was ended | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| Time of Last Successful Backup | Shows the period of time when the last successful backup process was ended | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| Storage Type | Shows the backup storage type | Default:<br/>Mode: absolute<br/>Level 1: 80<br/><br/> |  |
| Backup Last Version Size | Shows the size of the last backup |  |  |
| Storage Size/Free Space | Shows the space occupied by backups and the remaining amount of space. <br/>Not all storage supports total size, so "Total Volume Space" metrics can be zeroed |  |  |
| Time Spent on Backup | Shows time spent on last backup |  |  |
<!-- markdownlint-enable line-length -->
