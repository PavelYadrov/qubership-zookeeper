Automated testing
=================

1. pre-requisites:
   - [bash](https://en.wikipedia.org/wiki/Bash_(Unix_shell)) is available
   - [oc](https://github.com/openshift/origin/releases) (openshift-origin-client-tools) 
     is installed
   - three zookeeper PODs are running

2. clone repository
   ```
   git clone https://github.com/Netcracker/qubership-docker-zookeeper.git
   ```

3. `cd` to the directory
   ```
   cd docker-zookeeper/failover-scenarios/leader_node_deleted
   ```

4. make sure script file can be executed, i.e. do `chmod +x` on it:
   ```
   chmod +x test.sh
   ```

5. specify your zookeeper service name to `ZK_SERVICE_NAME` parameter in `test.sh` script
	
6. make sure that you have performed login and set active project for openshift client 
   to the project where zookeeper service is running

7. run `test.sh` script
   ```
   ./test.sh
   ```

You need to make sure new leader is elected.
