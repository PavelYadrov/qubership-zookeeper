# API Usage

For POST operations you must specify user/pass from `BACKUP_DAEMON_API_CREDENTIALS_USERNAME` and `BACKUP_DAEMON_API_CREDENTIALS_PASSWORD` env parameters so that you can use REST api to run backup tasks:

## Backup

### Full Manual Backup

If you want to make backup all ZooKeeper data, you need to run the following command:

```
curl -XPOST -u username:password http://localhost:8080/backup
```

After executing the command you receive name of folder where the backup is stored. For example,
`20190321T080000`.

### Backup Modes

If you want the backup to be performed in a certain mode, you need to specify variable `mode` in the
request body:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"mode":"transactional"}' http://localhost:8080/backup
```

For more information about `Backup Modes` see [Backup Modes](../backup-modes/backup-modes).

### Not Evictable Backup

If backup should not be evicted automatically, it is necessary to add `allow_eviction` property
with value `False` to the request body. For example,

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"allow_eviction":"False"}' http://localhost:8080/backup
```

### Backup Eviction

#### Evict Backup by ID

If you want to remove specific backup, you should run the following command:

```
curl -XPOST -u username:password http://localhost:8080/evict/<backup_id>
```

where `backup_id` is the name of necessary backup, e.g. `20190321T080000`. If operation is
successful, you see the following text: `Backup <backup_id> successfully removed`.

#### Evict Backups by Policy

If you need to remove all backups that match the specified **eviction policy**, run the following command:

```
curl -XPOST -u username:password http://localhost:8080/evict
```

Generally, eviction process is started after each successful backup, but you can run it manually.

### Backup Status

If backup is in progress, you can check its status running the following command:

```
curl -XGET http://localhost:8080/jobstatus/<backup_id>
```

where `backup_id` is backup name received at the backup execution step. The result is JSON with
the following information:

* `status` is status of operation, possible options: Successful, Queued, Processing, Failed
* `message` is description of error (optional field)
* `vault` is name of vault used in recovery
* `type` is type of operation, possible options: backup, restore
* `err` is last 5 lines of error logs if status=Failed, None otherwise
* `task_id` is identifier of the task

### Backup Information

To get the backup information, use the following command:

```
curl -XGET http://localhost:8080/listbackups/<backup_id>
```

where `backup_id` is the name of necessary backup. The command returns JSON string with data about
particular backup:

* `ts` is UNIX timestamp of backup
* `spent_time` is time spent on backup (in ms)
* `db_list` is list of stored databases
* `id` is backup name
* `size` is size of backup (in bytes)
* `evictable` is _true_ if backup is evictable, _false_ otherwise
* `locked` is _true_ if backup is locked (either process isn't finished, or it failed somehow)
* `exit_code` is exit code of backup script
* `failed` is _true_ if backup failed, _false_ otherwise
* `valid` is _true_ if backup is valid, _false_ otherwise

## Recovery

To recover data from certain backup you need to specify JSON with information about backup name
(`vault`) and databases (`dbs`). 
In case of the ZooKeeper databases are root znodes like `/zookeper`. Each database should be specified without
slashes, e.g. `zookeper`.

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"vault":"20190321T080000", "dbs":["zookeeper","tmp"]}' http://localhost:8080/restore
```

As a response you receive `task_id`, which can be used to check _Recovery Status_.

### Recovery Status

If recovery is in progress, you can check its status running the following command:

```
curl -XGET http://localhost:8080/jobstatus/<task_id>
```

where `task_id` is task id received at the recovery execution step.

## Backups List

To receive list of collected backups you need to use the following command:

```
curl -XGET http://localhost:8080/listbackups
```

It returns JSON with list of backup names.

## Backup Daemon Health

If you want to know the state of Backup Daemon, you should use the following command:

```
curl -XGET http://localhost:8080/health
```

As a result you receive JSON with information:

```
"status": status of backup daemon   
"backup_queue_size": backup daemon queue size (if > 0 then there are 1 or tasks waiting for execution)
 "storage": storage info:
  "total_space": total storage space in bytes
  "dump_count": number of backup
  "free_space": free space left in bytes
  "size": used space in bytes
  "total_inodes": total number of inodes on storage
  "free_inodes": free number of inodes on storage
  "used_inodes": used number of inodes on storage
  "last": last backup metrics
    "metrics['exit_code']": exit code of script 
    "metrics['exception']": python exception if backup failed
    "metrics['spent_time']": spent time
    "metrics['size']": backup size in bytes
    "failed": is failed or not
    "locked": is locked or not
    "id": vault name of backup
    "ts": timestamp of backup  
  "lastSuccessful": last succesfull backup metrics
    "metrics['exit_code']": exit code of script 
    "metrics['spent_time']": spent time
    "metrics['size']": backup size in bytes
    "failed": is failed or not
    "locked": is locked or not
    "id": vault name of backup
    "ts": timestamp of backup
```