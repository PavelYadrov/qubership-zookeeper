The following topics are covered in this chapter:

<!-- TOC -->
- [Single Node Failure](#single-node-failure)
  - [Description](#description)
  - [Alerts](#alerts)
  - [Stack Trace(s)](#stack-traces)
  - [How to Solve](#how-to-solve)
  - [Recommendations](#recommendations)
  
- [All Nodes Failure](#all-nodes-failure)
  - [Description](#description-1)
  - [Alerts](#alerts-1)
  - [Stack Trace(s)](#stack-traces-1)
  - [How to Solve](#how-to-solve-1)
  - [Recommendations](#recommendations-1)
  
- [CPU Limit](#cpu-limit)
  - [Description](#description-2)
  - [Alerts](#alerts-2)
  - [Stack Trace(s)](#stack-traces-2)
  - [How to Solve](#how-to-solve-2)
  - [Recommendations](#recommendations-2)
  
- [Memory Limit](#memory-limit)
  - [Description](#description-3)
  - [Alerts](#alerts-3)
  - [Stack Trace(s)](#stack-traces-3)
  - [How to Solve](#how-to-solve-3)
  - [Recommendations](#recommendations-3)
  
- [Disk Filled on All Nodes](#disk-filled-on-all-nodes)
  - [Description](#description-4)
  - [Alerts](#alerts-4)
  - [Stack Trace(s)](#stack-traces-4)
  - [How to Solve](#how-to-solve-4)
  - [Recommendations](#recommendations-4)
  
- [Last Backup Has Failed](#last-backup-has-failed)
  - [Description](#description-5)
  - [Alerts](#alerts-5)
  - [Stack Trace(s)](#stack-traces-5)
  - [How to Solve](#how-to-solve-5)
  - [Recommendations](#recommendations-5)
  
- [Data Is Out of Space](#data-is-out-of-space)
  - [Description](#description-6)
  - [Alerts](#alerts-6)
  - [Stack Trace(s)](#stack-traces-6)
  - [How to Solve](#how-to-solve-6)
  - [Recommendations](#recommendations-6)
  
- [Clients Cannot Connect To ZooKeeper with Connection Loss Error](#clients-cannot-connect-to-zookeeper-with-connection-loss-error)
  - [Description](#description-7)
  - [Alerts](#alerts-7)
  - [Stack Trace(s)](#stack-traces-7)
  - [How to Solve](#how-to-solve-7)
  - [Recommendations](#recommendations-7)
  
- [Container Failed with Error: container has runAsNonRoot and image will run as root](#container-failed-with-error-container-has-runasnonroot-and-image-will-run-as-root)
  - [Description](#description-8)
  - [Alerts](#alerts-8)
  - [Stack Trace(s)](#stack-traces-8)
  - [How to Solve](#how-to-solve-8)
  - [Recommendations](#recommendations-8)
  
- [ZooKeeper Backup Daemon Failed with PermissionError](#zookeeper-backup-daemon-failed-with-permissionerror)
  - [Description](#description-9)
  - [Alerts](#alerts-9)
  - [Stack Trace(s)](#stack-traces-9)
  - [How to Solve](#how-to-solve-9)
  - [Recommendations](#recommendations-9)
<!-- TOC -->

## Single Node Failure

### Description

A "Degraded" status on monitoring indicates that at least one of the nodes have failed.

_Leader node failure_\
ZooKeeper cluster will be temporary unable to process requests until new leader is elected.
Leader election is performed automatically by ZooKeeper cluster. For more information about
leader activation and election process in ZooKeeper please refer to ZooKeeper Documentation located at <https://zookeeper.apache.org/doc/r3.4.11/zookeeperInternals.html#sc_leaderElection>.
If ZooKeeper is unable to select new leader, check monitoring dashboard for other problems
that may prevent cluster from functioning correctly.

_Follower node failure_\
As long as a majority of the ensemble are up, ZooKeeper service will be available.

### Alerts

[ZooKeeper_Is_Degraded_Alarm](./alerts.md#ZooKeeper_Is_Degraded_Alarm)

### Stack trace(s)

```text
[WARN][zk_id=2][thread=QuorumPeer[myid=2](plain=/0.0.0.0:2181)(secure=disabled)][class=QuorumPeer@1287] Unexpected exception
java.lang.InterruptedException: Timeout while waiting for epoch from quorum
	at org.apache.zookeeper.server.quorum.Leader.getEpochToPropose(Leader.java:1227)
	at org.apache.zookeeper.server.quorum.Leader.lead(Leader.java:482)
	at org.apache.zookeeper.server.quorum.QuorumPeer.run(QuorumPeer.java:1284)
[INFO][zk_id=2][thread=QuorumPeer[myid=2](plain=/0.0.0.0:2181)(secure=disabled)][class=Leader@676] Shutting down
```

>**Note**
>If no specific stack trace is observed for the node failure, rely on monitoring logs to identify the issue.

### How to solve

Check the monitoring dashboard for any other problems that may have occurred around the same time the node failed.\
When the problem is localized, go to the appropriate problem description and troubleshooting procedure to fix it.\
Once the node is fixed, it will rejoin ensemble as a follower node.\
Try to reboot appropriate ZooKeeper Service deployments.

### Recommendations

Monitor node health using ZooKeeper Monitoring.\
Ensure redundancy to maintain majority availability.

## All Nodes Failure

### Description

A "Down" status on monitoring indicates that all of the ZooKeeper nodes have failed.

### Alerts

[ZooKeeper_Is_Down_Alarm](./alerts.md#ZooKeeper_Is_Down_Alarm)

### Stack trace(s)

```text
[WARN][zk_id=1][thread=NIOWorkerThread-2][class=NIOServerCnxn@373] Close of session 0x0
java.io.IOException: ZooKeeperServer not running
	at org.apache.zookeeper.server.NIOServerCnxn.readLength(NIOServerCnxn.java:544)
	at org.apache.zookeeper.server.NIOServerCnxn.doIO(NIOServerCnxn.java:332)
	at org.apache.zookeeper.server.NIOServerCnxnFactory$IOWorkRequest.doWork(NIOServerCnxnFactory.java:522)
	at org.apache.zookeeper.server.WorkerService$ScheduledWorkRequest.run(WorkerService.java:154)
	at java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1149)
	at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:624)
	at java.lang.Thread.run(Thread.java:748)
```

>**Note**
>If no specific stack trace is observed for the node failure, rely on monitoring logs to identify the issue.

### How to solve

Try to reboot all ZooKeeper Service deployments.

### Recommendations

Consider implementing health checks and redundancy measures to prevent complete cluster failures.

## CPU Limit

### Description

ZooKeeper request processing may be impacted up to potential node failure when
CPU consumption reaches resource limit for particular ZooKeeper node.

### Alerts

[ZooKeeper_CPU_Load_Alarm](./alerts.md#ZooKeeper_CPU_Load_Alarm)

### Stack trace(s)

```text
java.lang.OutOfMemoryError: unable to create new native thread
```

>**Note**
>If no specific stack trace is observed for the node failure, rely on monitoring logs to identify the issue.

### How to solve

Try to increase CPU request and CPU limit for appropriate ZooKeeper deployment.

### Recommendations

Review and update resource configurations to meet cluster demands.

## Memory Limit

### Description

ZooKeeper request processing may be impacted up to potential node failure when
memory consumption reaches resource limit.

### Alerts

[ZooKeeper_Memory_Usage_Alarm](./alerts.md#ZooKeeper_Memory_Usage_Alarm)

### Stack trace(s)

```text
java.lang.OutOfMemoryError: Java heap space
```

### How to solve

Try to increase Memory request and Memory limit for appropriate ZooKeeper deployment.\
`Note:` Memory request and Memory limit should be equal. If Memory request is changed, the `HEAP_OPTS`
environment variable of corresponding ZooKeeper deployment should be set as `New Memory request/2`.

For more detailed information, refer to [Memory Limit Guide](troubleshooting-scenarios/memory_limit.md).

### Recommendations

Monitor memory usage.\
Increase resource limits or scale out cluster in accordance.

## Disk Filled on All Nodes

### Description

ZooKeeper keeps all necessary data and clear the space according to `autopurge` configuration.
In case of incorrectly defined `autopurge` config or insufficient Disk Space, ZooKeeper nodes can be
totally filled that leads to unrecoverable crash of ZooKeeper.

### Alerts

Not applicable.

### Stack trace(s)

```text
[ERROR][zk_id=1][thread=QuorumPeer[myid=1]/0:0:0:0:0:0:0:0:2181][class=QuorumPeer@1395] Failed to write new file /var/opt/zookeeper/data/version-2/currentEpoch
java.io.IOException: No space left on device
	at java.io.FileOutputStream.writeBytes(Native Method)
    at java.io.FileOutputStream.write(FileOutputStream.java:326)
  	at org.apache.zookeeper.common.AtomicFileOutputStream.write(AtomicFileOutputStream.java:74)
  	at sun.nio.cs.StreamEncoder.writeBytes(StreamEncoder.java:221)
  	at sun.nio.cs.StreamEncoder.implFlushBuffer(StreamEncoder.java:291)
  	at sun.nio.cs.StreamEncoder.implFlush(StreamEncoder.java:295)
  	at sun.nio.cs.StreamEncoder.flush(StreamEncoder.java:141)
```

>**Note**
>If no specific stack trace is observed for the node failure, rely on monitoring logs to identify the issue.

### How to solve

Manually clean up the Disk space in case of high usage and adjust `autopurge` configuration.\
For more detailed information, refer to [Disk Filled Guide](troubleshooting-scenarios/disk_filled_on_all_nodes.md).

### Recommendations

Review disk usage and autopurge settings.

## Last Backup Has Failed

### Description

The last ZooKeeper backup has finished with `Failed` status.

### Alerts

[ZooKeeper_Last_Backup_Has_Failed_Alarm](./alerts.md#ZooKeeper_Last_Backup_Has_Failed_Alarm)

### Stack trace(s)

```text
OSError: [Error 28] No space left on device: '/opt/zookeeper/backup-storage/'
```

>**Note**
>If no specific stack trace is observed for the node failure, rely on monitoring logs to identify the issue.

### How to solve

Check that ZooKeeper Backup Daemon pod exists and is up. If ZooKeeper Backup Daemon is down, restart appropriate
deployment. If Backup Daemon pod is up, check it state by the following command from pod's terminal:

```sh
curl -XGET http://localhost:8080/health
```

### Recommendations

Not applicable.

## Data Is Out of Space

### Description

ZooKeeper becomes non-operational when disk capacity on a node runs out due to high volume of snapshot data and transactional log data.

### Alerts

Not applicable.

### Stack trace(s)

```text
[thread=QuorumPeer[myid=1](plain=/0:0:0:0:0:0:0:0:2181)(secure=disabled)][class=FileSnap@83] Reading snapshot /var/opt/zookeeper/data/version-2/snapshot.1d0000036d
[ERROR][zk_id=1][thread=QuorumPeer[myid=1](plain=/0:0:0:0:0:0:0:0:2181)(secure=disabled)][class=QuorumPeer@955] Unable to load database on disk
```

### How to solve

You may need to manually clean up this data occasionally. For this purpose you should find ZooKeeper folder
`\var\opt\zookeeper\data` and delete `version-2` folder using bash-command `rm -rf version-2`.

### Recommendations

Ensure sufficient disk capacity for ZooKeeper nodes based on the expected data growth and retention policies.\
Review disk usage.

## Clients Cannot Connect To ZooKeeper with Connection Loss Error

### Description

ZooKeeper clients cannot connect to ZooKeeper server with `Connection Loss` errors.\
It could happen after restart ZooKeeper leader (during upgrade, change configuration, failover scenarios, etc.) due to incorrect shutdown.\
There is external ticket for this issue: <https://issues.apache.org/jira/browse/ZOOKEEPER-3828> but there is no solution.

>**NOTE:** Starting with ZooKeeper `3.5.8-1.1` the ZooKeeper cluster is marked as not ready if clients cannot connect to it.
Before this version ZooKeeper cluster looks like fully operational even if clients cannot connect to it.

### Alerts

Not applicable.

### Stack trace(s)

```text
<Error> prom.samples (ReplicatedMergeTreeRestartingThread): void DB::ReplicatedMergeTreeRestartingThread::run(): Code: 999, e.displayText() = Coordination::Exception: All connection tries failed while connecting to ZooKeeper. nodes: 172.30.161.184:2181
Code: 209, e.displayText() = DB::NetException: Timeout exceeded while reading from socket 
172.30.161.184:2181
(Connection loss), Stack trace (when copying this message, always include the lines below):
```

or

```text
[2020-08-06T12:47:06,278][INFO][zk_id=localhost:2181][thread=main-SendThread(localhost:2181)][class=ClientCnxn$SendThread@1238] Client session timed out, have not heard from server in 30004ms for s
essionid 0x0, closing socket connection and attempting reconnect
KeeperErrorCode = ConnectionLoss for                                                                                                                                      
```

while ZooKeeper pods are run and do not contain any problems in logs.

### How to solve

You need to completely restart ZooKeeper cluster:
scale down all ZooKeeper Deployment Configs and when all the ZooKeeper pods are removed scale up Deployment Configs.

In case if this solution did not help you need to manually kill ZooKeeper process on each ZooKeeper pod.
For this perform the following command in terminal of each ZooKeeper pod:

```bash
kill -9 $(pidof java)
```

_Permanent Solution_

* Upgrade ZooKeeper to 3.6.2.
* If the issue is reproduced in OpenShift 3.11 environment it is strongly recommended to upgrade OpenShift Installer build to 2.65.

### Recommendations

How to check it is the same problem:

1. Try to execute the following command from any ZooKeeper's pod terminal:

    ```bash
    ./bin/zkCli.sh ls /
    ```

2. If ZooKeeper Cli does not return list of znodes and after some time returns `ConnectionLoss` error it means that this is a problem from this topic.

## Container Failed with Error: container has runAsNonRoot and image will run as root

### Description

The Operator deploys successfully and operator logs do not contain errors, but ZooKeeper Monitoring and/or ZooKeeper Backup Daemon pods fail.

ZooKeeper Monitoring and ZooKeeper Backup Daemon do not have special user to run processes, so default (`root`) user is used.
If you miss the `securityContext` parameter in the pod configuration and `Pod Security Policy` is enabled, the default `securityContext` for pod is taken from `Pod Security Policy`.
If you configure the `Pod Security Policy` as follows then the error mentioned above occurs:

```yaml
runAsUser:
  # Require the container to run without root privileges.
  rule: 'MustRunAsNonRoot'
```

### Alerts

Not applicable.

### Stack trace(s)

```text
Error: container has runAsNonRoot and image will run as root
```

### How to solve

You need to specify the correct `securityContext` in pod configuration during installation. For example, for ZooKeeper Monitoring and ZooKeeper Backup Daemon, you should specify the following parameter:

```yaml
securityContext:
    runAsUser: 1000
```

### Recommendations

Not applicable.

## ZooKeeper Backup Daemon Failed with PermissionError

### Description

The Operator deploys successfully and operator logs do not contain errors, but ZooKeeper Backup Daemon fails.

### Alerts

Not applicable.

### Stack trace(s)

```text
[INFO] Init storage object with storage root: /opt/zookeeper/backup-storage
Traceback (most recent call last):
  File "/opt/backup/backup-daemon.py", line 506, in <module>
    backup_processor = BackupProcessor(backup_cfg, FULL)
  File "/opt/backup/backup-daemon.py", line 28, in __init__
    self.storage = storage.Storage(proc_config['storage_root'])
  File "/opt/backup/storage.py", line 35, in __init__
    os.makedirs(self.granular_folder)
  File "/usr/local/lib/python3.7/os.py", line 221, in makedirs
    mkdir(name, mode)
PermissionError: [Errno 13] Permission denied: '/opt/zookeeper/backup-storage/granular'/ code placeholder
```

### How to solve

To fix the problem with lack of rights to persistent volume, specify `securityContext` with `fsGroup` in the pod configuration during installation:

```yaml
securityContext:
    fsGroup: 1000
```

### Recommendations

Not applicable.
