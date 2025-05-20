This chapter describes the security audit logging for ZooKeeper.

<!-- #GFCFilterMarkerStart# -->
The following topics are covered in this chapter:
<!-- #GFCFilterMarkerEnd# -->

# Common Information

Audit logs let you track access to your ZooKeeper cluster and are useful for compliance purposes or in the aftermath of a security breach.
You can find more detailed information about audit logs and their configuration in official documentation [ZooKeeper Audit Logs](https://zookeeper.apache.org/doc/r3.8.4/zookeeperAuditLogs.html).

# Configuration

To enable ZooKeeper audit logs need to set `zookeeper.auditEnabled` parameter to `true`:

```yaml
zookeeper:
  auditEnabled: true
```

By default, audit logging for ZooKeeper is disabled.

## Example of Events

The audit log format for events are described further:

### Create Session

A session successfully created.

```text
2024-08-14 15:24:22,269 [myid:] - INFO  [NIOWorkerThread-1:o.a.z.s.ZooKeeperServer@1653] - Session 0x100da503a060000: auth success for scheme digest and address /127.0.0.1:54780


2024-08-14 15:23:42,065 [myid:] - INFO  [QuorumConnectionThread-[myid=1]-3:o.a.z.s.q.a.SaslQuorumServerCallbackHandler@163] - Successfully authenticated learner: authenticationID=zadmin;  authorizationID=zadmin.
```

### Close Session

A session successfully closed.

```text
2024-08-14 15:24:22,365 [myid:] - INFO  [RequestThrottler:o.a.z.s.q.QuorumZooKeeperServer@163] - Submitting global closeSession request for session 0x100da503a060000
```

### Unauthenticated event

The user authentication failed.

```text
2024-08-14 15:28:56,150 [myid:] - WARN  [NIOWorkerThread-1:o.a.z.s.ZooKeeperServer@1751] - Client /0:0:0:0:0:0:0:1:56320 failed to SASL authenticate: {}
```

### Unauthorized event

The user does not have the required permissions to make the request.

```text
2024-08-14 15:26:10,192 [myid:] - ERROR [CommitProcessor:1:o.a.z.a.Slf4jAuditLogger@33] - session=0x100da503a060005    user=zadmin,127.0.0.1    ip=127.0.0.1    operation=create    znode=/test    znode_type=persistent    result=failure
```
