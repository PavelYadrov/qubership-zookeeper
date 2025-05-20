[[_TOC_]]

# Frequently Asked Questions

## What is ZooKeeper?

ZooKeeper is a centralized service for maintaining configuration information, naming,
providing distributed synchronization, and providing group services.
All of these kinds of services are used in some form or another by distributed applications.

## Official sources

| Name              | Type    | Language | Link                                                                |
|-------------------|---------|----------|---------------------------------------------------------------------|
| Web-site          | Article | EN       | [Link](https://zookeeper.apache.org/)                               |
| Documentation     | Article | EN       | [Link](https://zookeeper.apache.org/doc/current/index.html)         |
| Overview          | Article | EN       | [Link](https://zookeeper.apache.org/doc/current/zookeeperOver.html) |
| ZooKeeper Curator | Article | EN       | [Link](https://curator.apache.org/docs/about)                       |

## External sources

| Name                 | Type    | Language | About                                                                                        | Link                                                                                                                           |
|----------------------|---------|----------|----------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| "What is Zookeeper?" | video   | EN       | Briefly ZooKeeper description                                                                | [YouTube](https://youtu.be/AS5a91DOmks?si=uDf3_ZkQbWGUY9cf)                                                                    |
| "Apache ZooKeeper"   | video   | RU       | ZooKeeper idea, architecture and code examples                                               | [YouTube.p1](https://youtu.be/PgTpvzv8xp0?si=T2afqCQOChjRcbwq), [YouTube.p2](https://youtu.be/hYDrkFjznHQ?si=IuDRf5cW04iklx45) |
| "What is Zookeeper?" | article | EN       | Light version of official documentation overview, also with code examples on Java and Python | [GeeksForGeeks](https://www.geeksforgeeks.org/what-is-apache-zookeeper/)                                                       |

## Developer hints

### Working with ZooKeeper via zkCli.sh

If you would like to check ZooKeeper content or do some changes, then you can use zkCli.sh in ZooKeeper pod:

1. Run terminal in a ZooKeeper pod.
2. Run `./bin/zkCli.sh`.

If ZooKeeper is secured, then you need to specify creds. For more info visit [Security Guide](../public/security.md).

After that you will working in specific ZooKeeper terminal where you can go through zNodes and work with them.

You can read more in [Connecting to ZooKeeper article](https://zookeeper.apache.org/doc/current/zookeeperStarted.html#sc_ConnectingToZooKeeper).

### Working with ZooKeeper via UI

We also have charts for deploying ZooKeeper UI.

**Attention**: it can be used just for internal development work.
It is not presented on dev/qa/project cluster, but you can deploy it using HELM only in your own dev-environment.

### ZooKeeper JVM parameters check

#### All

1. Go in one of the ZooKeeper container.

2. Find running ZooKeeper process by using the following command: `ps aux`.
 You need to find `process_id` for command which is starting with `/opt/java/openjdk/bin/java...`.

3. Print system properties:

   ```sh
   jattach <process_id> jcmd "VM.system_properties"
   ```

4. See all availability of jattach:

   ```sh
   jattach <process_id> jcmd "help -all" 
   ```

#### Heap size

1. The same as in [All case](#all)

2. The same as in [All case](#all)

3. Run:

  ```sh
  jattach <pid> jcmd "VM.command_line"
  jattach <pid> jcmd "VM.flags"
  ```
