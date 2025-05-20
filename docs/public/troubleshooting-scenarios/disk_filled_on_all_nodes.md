# Disk Filled on All Nodes

## Problem Detection

### Metric

You can use information about mounted volumes that is stored in `t_disk` measurement with fields:

- `used`
- `used_percent`

To retrieve information about how much disk space ZooKeeper is using:

- Run the command below inside ZooKeeper container.

  ```sh
  df -h /var/opt/zookeeper/data
  ```

Possible output:

```text
Filesystem      Size  Used Avail Use% Mounted on                                                                                                                                                       
/dev/vdc        2.0G  6.5M  1.8G   1% /var/opt/zookeeper/data
```

### Grafana Dashboard

To retrieve the metric of disk usage the following query can be used:

```text
SELECT mean("used_percent") FROM "t_disk" WHERE ("path" =~ /.*$zk_pv_name/) AND $timeFilter GROUP BY time($inter), "path" fill(null)
```

You should check in the drop-down menu `zk_pv_name` in Grafana and on the `Disk usage` (and `Disk usage in percent`) graph that necessary PV names are selected.

### Logging

When the problem occurs, you can see the following exception in the console logs:

```text
[2018-03-14T14:26:11,411][ERROR][zk_id=1][thread=QuorumPeer[myid=1]/0:0:0:0:0:0:0:0:2181][class=QuorumPeer@1395] Failed to write new file /var/opt/zookeeper/data/version-2/currentEpoch
java.io.IOException: No space left on device
	at java.io.FileOutputStream.writeBytes(Native Method)
    at java.io.FileOutputStream.write(FileOutputStream.java:326)
  	at org.apache.zookeeper.common.AtomicFileOutputStream.write(AtomicFileOutputStream.java:74)
  	at sun.nio.cs.StreamEncoder.writeBytes(StreamEncoder.java:221)
  	at sun.nio.cs.StreamEncoder.implFlushBuffer(StreamEncoder.java:291)
  	at sun.nio.cs.StreamEncoder.implFlush(StreamEncoder.java:295)
  	at sun.nio.cs.StreamEncoder.flush(StreamEncoder.java:141)
  	at java.io.OutputStreamWriter.flush(OutputStreamWriter.java:229)
  	at java.io.BufferedWriter.flush(BufferedWriter.java:254)
  	at org.apache.zookeeper.server.quorum.QuorumPeer.writeLongToFile(QuorumPeer.java:1391)
  	at org.apache.zookeeper.server.quorum.QuorumPeer.setCurrentEpoch(QuorumPeer.java:1426)
  	at org.apache.zookeeper.server.quorum.Learner.syncWithLeader(Learner.java:454)
  	at org.apache.zookeeper.server.quorum.Follower.followLeader(Follower.java:83)
  	at org.apache.zookeeper.server.quorum.QuorumPeer.run(QuorumPeer.java:981)
```

## Troubleshooting Procedure

ZooKeeper contains files that are persistent copies of znodes stored as snapshots and transactional
log files. When changes are made to znodes, these changes are appended to the transactional log,
and, eventually, a snapshot of the current state of all znodes is written to the file system.

By default, ZooKeeper enables automatic purging of snapshots and corresponding transaction logs:

```text
# The number of snapshots to retain in dataDir
autopurge.snapRetainCount=3
# The time interval in hours for which the purge task has to be triggered
autopurge.purgeInterval=6
```

For more information, refer to Maintenance documentation located at <http://zookeeper.apache.org/doc/current/zookeeperAdmin.html#sc_maintenance>.

Still we recommend to watch for disk usage and to execute the following commands to clear snapshots
and logs when disk is > 85% usage:

```sh
cd /opt/zookeeper/bin
./zkCleanup.sh /var/opt/zookeeper/data -n 3
```

Because ZooKeeper doesn't recover from crash when disk was full:

```text
[2018-03-13T15:40:23,509][ERROR][zk_id=][thread=main][class=ZooKeeperServerMain@66] Unexpected exception, exiting abnormally
java.io.EOFException
	  at java.io.DataInputStream.readInt(DataInputStream.java:392)
	  at org.apache.jute.BinaryInputArchive.readInt(BinaryInputArchive.java:63)
	  at org.apache.zookeeper.server.persistence.FileHeader.deserialize(FileHeader.java:66)
	  at org.apache.zookeeper.server.persistence.FileTxnLog$FileTxnIterator.inStreamCreated(FileTxnLog.java:589)
	  at org.apache.zookeeper.server.persistence.FileTxnLog$FileTxnIterator.createInputArchive(FileTxnLog.java:608)
	  at org.apache.zookeeper.server.persistence.FileTxnLog$FileTxnIterator.goToNextLog(FileTxnLog.java:574)
  	at org.apache.zookeeper.server.persistence.FileTxnLog$FileTxnIterator.next(FileTxnLog.java:654)
  	at org.apache.zookeeper.server.persistence.FileTxnSnapLog.restore(FileTxnSnapLog.java:166)
  	at org.apache.zookeeper.server.ZKDatabase.loadDataBase(ZKDatabase.java:223)
  	at org.apache.zookeeper.server.ZooKeeperServer.loadData(ZooKeeperServer.java:283)
  	at org.apache.zookeeper.server.ZooKeeperServer.startdata(ZooKeeperServer.java:406)
  	at org.apache.zookeeper.server.NIOServerCnxnFactory.startup(NIOServerCnxnFactory.java:118)
  	at org.apache.zookeeper.server.ZooKeeperServerMain.runFromConfig(ZooKeeperServerMain.java:121)
	  at org.apache.zookeeper.server.ZooKeeperServerMain.initializeAndRun(ZooKeeperServerMain.java:89)
	  at org.apache.zookeeper.server.ZooKeeperServerMain.main(ZooKeeperServerMain.java:55)
  	at org.apache.zookeeper.server.quorum.QuorumPeerMain.initializeAndRun(QuorumPeerMain.java:119)
	  at org.apache.zookeeper.server.quorum.QuorumPeerMain.main(QuorumPeerMain.java:81)
```
