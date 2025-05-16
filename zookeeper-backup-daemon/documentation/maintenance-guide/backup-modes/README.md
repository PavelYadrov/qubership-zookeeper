# Backup and Recovery

This section describes the backup and recovery procedures for ZooKeeper in detail.

## Backup Modes

There are two modes to backup ZooKeeper data:

* `Transactional` backup represents the saving of the snapshot and transaction logs that ZooKeeper
  leader generates.
* `Hierarchical` backup represents the preservation of the complete structure of the ZooKeeper nodes.

There is ability to specify which mode will be used. The user needs to
specify variable `mode` in request body

```
curl -XPOST -v -H "Content-Type: application/json" -d '{"mode":"transactional"}' http://localhost:8080/backup
```

The variable `mode` can be one of two values: `transactional` and `hierarchical`, which are responsible
for `transactional` and `hierarchical` backup respectively. If `mode` is not set, `hierarchical` mode will be used by default.

### Transactional backup

This mode of backup *requires the shared file system* because it uses snapshots and transaction logs
that are created by ZooKeeper itself and stored in its file system. The snapshots and transaction
logs are taken from the ZooKeeper leader pod because it stores the most actual data.

At the REST request from `ZooKeeper Backup Daemon` ZooKeeper leader pod transfers all data from
inner file system (directory `/var/opt/zookeeper/data/version-2`) to shared file system (directory
`/opt/zookeeper/backup-storage/tmp`). After moving data `ZooKeeper Backup Daemon` chooses which
logs are used for backup creation: the last snapshot and transaction logs that are collected after
creation of the last snapshot. The last snapshot is copied to the directory for particular backup
`/opt/zookeeper/backup-storage/<backup_id>`. In the same directory filtered transaction logs are
stored. Logs filtering means deleting all ephemeral nodes and sessions transactions, the remaining
transactions are saved in the file with the same name. After successful backup the temporary
directory `/opt/zookeeper/backup-storage/tmp` is deleted.

This backup mode is *consistent* because ZooKeeper structure is saved in an instant by copying
necessary logs.

**Important:** Transactional backup does not work in DR mode with joint ZooKeeper cluster.

### Hierarchical backup

With this mode of backup `ZooKeeper Backup Daemon` connects to ZooKeeper using the client and saves structure of znodes
hierarchically to a directory for particular backup `/opt/zookeeper/backup-storage/<backup_id>`.
Each znode is saved as a directory with znode name and znode data is stored in a `content` file
inside the directory for particular znode. After that backup daemon creates archive with all directories inside.

This backup mode is *not consistent* because ZooKeeper structure preservation can be accompanied by
the znode creation, modification or deletion.

**NOTE:** Hierarchical granular backup/restore are enabled only for root znodes recursively with all znode children and their content.

### Transactional Restore

Recovery from `transactional` backup *requires restarting all ZooKeeper pods*.

`ZooKeeper Backup Daemon` starts recovery from transferring backup data from the directory with
particular backup `/opt/zookeeper/backup-storage/<backup_id>` to the temporary directory
`/opt/zookeeper/backup-storage/recover`. Then ZooKeeper service forcibly restarts. During restart
each ZooKeeper pod substitutes snapshots and transaction logs in directory
`/var/opt/zookeeper/data/version-2` with data from directory `/opt/zookeeper/backup-storage/recover`.
After successful backup the temporary directory `/opt/zookeeper/backup-storage/recover` is deleted.

This recovery allows to restore the system state that was at a moment of the backup (full recovery).

**NOTE:** Parameter `dbs` should not be used for recovering from transactional backup.

### Hierarchical Restore

Recovery from `hierarchical` backup *does not require restarting the server*, because the client API
is used to restore znode tree.

`ZooKeeper Backup Daemon` connects to ZooKeeper using the client and starts to visit znodes in
hierarchy from backup directory `/opt/zookeeper/backup-storage/<backup_id>`. `ZooKeeper Backup Daemon`
traverses subtree of root znodes and recover znodes that are absent in ZooKeeper structure, exisitng znodes will be replaced with znodes from backup.