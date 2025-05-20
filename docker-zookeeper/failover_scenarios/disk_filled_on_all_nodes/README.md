Automated testing
=================

1. pre-requisites:
   - [bash](https://en.wikipedia.org/wiki/Bash_(Unix_shell)) is available
   - [oc](https://github.com/openshift/origin/releases) (openshift-origin-client-tools) 
     is installed
   - [jq](https://stedolan.github.io/jq/) is installed
   - you have persistent volume with capacity 2GB
   - one ZooKeeper POD is running

2. clone repository
   ```
   git clone https://github.com/Netcracker/qubership-docker-zookeeper.git
   ```

3. `cd` to the directory
   ```
   cd docker-zookeeper/failover-scenarios/disk_filled_on_all_nodes
   ```

4. make sure script file can be executed, i.e. do `chmod +x` on it:
   ```
   chmod +x collect_metrics.sh
   chmod +x test.sh
   ```

5. specify your ZooKeeper pod name to `ZOOKEEPER_POD_NAME` parameter in `test.sh` and 
   `collect_metrics.sh` scripts

6. specify your OpenShift server URL to `OPENSHIFT_URL` parameter in `test.sh` script

7. specify your name of OpenShift project to `PROJECT_NAME` parameter in `test.sh` script

8. run `test.sh` script
   ```
   ./test.sh
   ```

9. run `collect_metrics.sh` script for monitoring disk usage. It must be run in parallel with 
   `test.sh` script
   ```
   ./collect_metrics.sh
   ```

You need to make sure test is failed with exception:
```
Exception in thread "main" org.apache.zookeeper.KeeperException$ConnectionLossException: KeeperErrorCode = ConnectionLoss for /test_failover_scenarios/disk_filled_on_all_nodes/1
        at org.apache.zookeeper.KeeperException.create(KeeperException.java:102)
        at org.apache.zookeeper.KeeperException.create(KeeperException.java:54)
        at org.apache.zookeeper.ZooKeeper.create(ZooKeeper.java:786)
        at org.apache.zookeeper.ZooKeeperMain.processZKCmd(ZooKeeperMain.java:707)
        at org.apache.zookeeper.ZooKeeperMain.processCmd(ZooKeeperMain.java:600)
        at org.apache.zookeeper.ZooKeeperMain.run(ZooKeeperMain.java:363)
        at org.apache.zookeeper.ZooKeeperMain.main(ZooKeeperMain.java:291)
```

After script execution you can clear disk space using command:
```
oc exec ${ZOOKEEPER_POD_NAME} -- rm /var/opt/zookeeper/data/busy_space
oc exec ${ZOOKEEPER_POD_NAME} -- ./bin/zkCli.sh rmr /test_failover_scenarios
```
