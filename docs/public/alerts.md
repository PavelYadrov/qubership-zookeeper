# Prometheus Alerts

## ZooKeeper_Is_Degraded_Alarm

### Description

ZooKeeper cluster degraded, it means that at least one of the nodes have failed, but cluster is able to work.

More information in [Single Node Failure](./troubleshooting.md#single-node-failure).

### Possible Causes

- ZooKeeper pod failures or unavailability.
- Resource constraints impacting ZooKeeper pod performance.

### Impact

- Reduced or disrupted functionality of the ZooKeeper cluster.
- Potential impact on Kafka, Config Server, Discovery and other processes relying on the ZooKeeper.

### Actions for Investigation

1. Check the status of ZooKeeper pods.
2. Review logs for ZooKeeper pods for any errors or issues.
3. Verify resource utilization of ZooKeeper pods (CPU, memory).

### Recommended Actions to Resolve Issue

1. Restart or redeploy ZooKeeper pods if they are in a failed state.
2. Investigate and address any resource constraints affecting the ZooKeeper pod performance.

## ZooKeeper_Is_Down_Alarm

### Description

ZooKeeper cluster is down, and there are no available pods.

More information in [All Nodes Failure](./troubleshooting.md#all-nodes-failure).

### Possible Causes

- Network issues affecting the ZooKeeper pod communication.
- ZooKeeper's storage is corrupted.
- Internal error blocks ZooKeeper cluster working.

### Impact

- Complete unavailability of the ZooKeeper cluster.
- Kafka, Config Server, Discovery and other processes relying on the ZooKeeper will fail.

### Actions for Investigation

1. Check the status of ZooKeeper pods.
2. Review logs for ZooKeeper pods for any errors or issues.
3. Verify resource utilization of ZooKeeper pods (CPU, memory).

### Recommended Actions to Resolve Issue

1. Check the network connectivity to the ZooKeeper pods.
2. Check the ZooKeeper storage for free space or data corruption.
3. Restart or redeploy all ZooKeeper pods once.

## ZooKeeper_CPU_Load_Alarm

### Description

One of ZooKeeper pods uses 95% of the CPU limit.

More information in [CPU Limit](./troubleshooting.md#cpu-limit).

### Possible Causes

- Insufficient CPU resources allocated to ZooKeeper pods.
- The service is overloaded.

### Impact

- Increased response time and potential slowdown of ZooKeeper requests.
- Degraded performance of services used the ZooKeeper.

### Actions for Investigation

1. Monitor the CPU usage trends ZooKeeper Monitoring dashboard.
2. Review ZooKeeper logs for any performance related issues.

### Recommended Actions to Resolve Issue

1. Try to increase CPU request and CPU limit for appropriate ZooKeeper deployment.
2. Scale up ZooKeeper cluster if needed.

## ZooKeeper_Memory_Usage_Alarm

### Description

One of ZooKeeper pods uses 90% of the memory limit.

More information in [Memory Limit](./troubleshooting.md#memory-limit).

### Possible Causes

- Insufficient memory resources allocated to ZooKeeper pods.
- Memory leaks or excessive memory ephemeral ZNodes on the ZooKeeper side.

### Impact

- Potential out-of-memory errors and ZooKeeper cluster instability.
- Degraded performance of services used the ZooKeeper.

### Actions for Investigation

1. Monitor the Memory usage trends ZooKeeper Monitoring dashboard.
2. Review ZooKeeper logs for memory related errors.

### Recommended Actions to Resolve Issue

1. Try to increase Memory request, Memory limit and Heap Size for appropriate ZooKeeper deployment.
2. Scale up ZooKeeper cluster if needed.

## ZooKeeper_Last_Backup_Has_Failed_Alarm

### Description

The last ZooKeeper backup has finished with `Failed` status.

More information in [Last Backup Has Failed](./troubleshooting.md#last-backup-has-failed).

### Possible Causes

- Unavailable or broken backup storage (Persistent Volume or S3).
- Network issues affecting the ZooKeeper and Backup Daemon pod communication.

### Impact

- Not available backup for ZooKeeper and no ability to restore it in case of disaster.

### Actions for Investigation

1. Monitor the Backup Daemon state on Backup Daemon Monitoring dashboard.
2. Review ZooKeeper Backup Daemon logs for investigation of cases the issue.
3. Check backup storage.

### Recommended Actions to Resolve Issue

1. Fix issues with backup storage if necessary.
2. Follow [Last Backup Has Failed](./troubleshooting.md#last-backup-has-failed) for additional steps.
