## Stress Testing

This chapter describes the effect of Zookeeper Backup Daemon on ZooKeeper performance.

### How To

1. Run performance test of ZooKeeper
2. When `mixed` step is started, run backup using the command below:
   ```
   curl -XPOST http://localhost:8080/backup
   ```
   where `localhost:8080` is host of Zookeeper Backup Daemon.

### Test Result

| cases                | znode count | throughput, msg/sec |
| -------------------- | ----------- | ------------------- |
| only test load       | 40149       | 550.67              |
| test load and backup | 35588       | 525.75              |

Result:
* throughput, diff = (525.75 - 550.67) / 525.75 = -5%
