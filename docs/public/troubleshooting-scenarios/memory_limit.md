# Memory Limit

## Problem Detection

### Metric

You can use information about used memory that is stored in `memory/usage`
measurement in field `value` and information about memory limit that is stored
in `memory/limit` measurement in field `value`.

These metrics show pod memory usage and limit in bytes. Constant high memory usage
that is close to memory limit may indicate a critical situation that service is
overloaded or resource limits are too low. It can potentially lead to the
increase of response times or crashes.

### Grafana Dashboard

To retrieve the metric of memory usage the following query can be used:

```text
SELECT mean("value") FROM "memory/usage" WHERE ("pod_namespace" =~ /^$project$/ AND "type" = 'pod' AND "pod_name" =~ /zookeeper(.*)/) AND $timeFilter GROUP BY time($inter), "pod_name" fill(linear)
```

To retrieve the metric of memory limit the following query can be used:

```text
SELECT mean("value") FROM "memory/limit" WHERE ("pod_namespace" =~ /^$project$/ AND "type" = 'pod' AND "pod_name" =~ /zookeeper(.*)/) AND $timeFilter GROUP BY time($inter), "pod_name" fill(linear)
```

## Troubleshooting Procedure

If you see a high memory usage, you can either increase memory limit along with
heap size or scale out the cluster by adding more nodes. It is also possible
update client throttling configuration in ZooKeeper by setting lower value to
globalOutstandingLimit property in zookeeper. It will reduce amount of outstanding
requests in the system. New value can be passed in `SERVER_JVMFLAGS` environment
variable on deployment configuration like this:

```text
-Dzookeeper.globalOutstandingLimit=800
```
