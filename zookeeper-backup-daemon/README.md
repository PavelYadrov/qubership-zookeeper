Zookeeper Backup Daemon
=======================

# Manual Backup Procedure

To perform manual backup:

*Find zookeeper leader
  * Go to Openshift: **Projects -> zookeeper-service -> pods -> zookeeper pod. 
  * Run "bin/zkServer.sh status" If you found "Mode: leader" then this pod is leader
* Go to leader PV
* Create backup archive: zip backup.zip snapshot.* log.*

# Manual Restore Procedure

To perform manual restore:

* Stop all zookeeper nodes 
* Clean up version-2 folder in all zookeepers PV   
* Choose any PV and extract all files from backup to it. unzip -oq backup.zip

