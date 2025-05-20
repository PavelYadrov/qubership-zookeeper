# What is Apache ZooKeeper?

Apache ZooKeeper is a software project of the Apache Software Foundation, providing an open source 
distributed configuration service, synchronization service, and naming registry for large 
distributed systems. ZooKeeper was a sub-project of Hadoop but is now a top-level 
project in its own right.

Reliability is provided by clustering multiple ZooKeeper processes, and since ZooKeeper uses quorums 
you need an odd number (typically 3 or 5 in a production environment).


# How to use this image

This image can be used to run one or more instances of ZooKeeper. If running a single instance, 
the defaults are often good enough, especially for simple evaluations and demonstrations. However, 
when running multiple instances you will need to use the environment variables.

## docker-compose.yml

```yaml
zookeeper:
  restart: always
  image: zookeeper:3.8.3
  ports:
    - "2181:2181"
  volumes:
    - ./config/zookeeper/zookeeper.properties:/var/opt/zookeeper/zookeeper.properties
    - ./config/zookeeper/zookeeper_jaas.conf:/var/opt/zookeeper/zookeeper_jaas.conf
    - /srv/docker/zookeeper/data/:/var/opt/zookeeper/zookeeper-data/
  environment:
    SERVER_JVMFLAGS: '-Djava.security.auth.login.config=/var/opt/zookeeper/zookeeper_jaas.conf'
```


# Environment variables

### `SERVER_NAME`

This environment variable defines part of ZooKeeper server name. The default is 'zookeeper'.

### `SERVER_ID`

This environment variable defines the numeric identifier for this ZooKeeper server. The default is 
'1' and is only applicable for a single standalone ZooKeeper server that is not replicated or fault 
tolerant. In all other cases, you should set the server number to a unique value within your 
ZooKeeper cluster.

### `SERVER_COUNT`

This environment variable defines the total number of ZooKeeper servers in the cluster. The default 
is '1' and is only applicable for a single standalone ZooKeeper server. In all other cases, you must 
use this variable to set the total number of servers in the cluster.

### `IS_CLOUD_ENABLED`

This environment variable defines the mode in which Zookeeper quorum will run. The default is 
'true'. If 'false' the quorum server list must be specified in `QUORUM_SERVERS`.

### `QUORUM_SERVERS`

This environment variable defines the list of quorum servers. The number of servers must be the same
as `SERVER_COUNT`. A comma must be used as a separator. Format of each node: 
`<server_id>@<zookeeper_host>:<followers_port>:<election_port>:[<role>]`.

### `LOG_LEVEL`

This environment variable is optional. Use this to set the level of detail for ZooKeeper's 
application log written to STDOUT and STDERR. Valid values are `INFO` (default), `WARN`, `ERROR`, 
`DEBUG`, or `TRACE`."

### `JOLOKIA_PORT` 

The number of port for Jolokia JMX agent. If empty the Jolokia agent is not started.
This feature is **deprecated** and will be removed in further releases.


# Ports

Containers created using this image will expose ports 2181, 2888, and 3888. These are the standard 
ports used by ZooKeeper. You can  use standard Docker options to map these to different ports on 
the host that runs the container.


# Storing data

The ZooKeeper run by this image writes data to the local file system, and the only way to keep this 
data is to volumes that map specific directories inside the container to the local file system 
(or to OpenShift persistent volumes).

### Zookeeper data

This image defines a data volume at `/var/opt/zookeeper/data`, and it is in this directory that 
the Zookeeper server will persist all of its data. You must mount it appropriately when running your 
container to persist the data after the container is stopped; failing to do so will result in all 
data being lost when the container is stopped.
