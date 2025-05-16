# ZooKeeper Monitoring Demo

To run demo execute the following command:
```
docker-compose up
```

It will start the ZooKeeper cluster with 3 containers and container with zookeeper monitoring.

Prometheus endpoint will be available here: http://localhost:8096/metrics

You can find Grafana dashboard for Prometheus metrics here: https://github.com/Netcracker/qubership-zookeeper/blob/main/zookeeper-service-operator/charts/helm/zookeeper-service/monitoring/zookeeper-dashboard.json

# ZooKeeper Monitoring Parameters

* `ZOOKEEPER_HOST` - The comma-separated list of zookeeper server connections to monitor. For example, `'zookeeper-1:2181','zookeeper-2:2181','zookeeper-3:2181'`. IP and DNS names can be used.
* `PROMETHEUS_URLS`- The comma-separated list of zookeeper prometheus URLs to obtain metrics. For example, `'http://zookeeper-1:8080/metrics','http://zookeeper-2:8080/metrics','http://zookeeper-3:8080/metrics'`. IP and DNS names can be used. If empty prometheus URLs are automatically generated from `ZOOKEEPER_HOST`with `8080` port.
* `ZOOKEEPER_CLIENT_USERNAME` - The username of user account to connect to ZooKeeper cluster. Should be empty if ZooKeeper is not secured.
* `ZOOKEEPER_CLIENT_PASSWORD` - The password of user account to connect to ZooKeeper cluster. Should be empty if ZooKeeper is not secured.
* `OS_PROJECT` - The value of label `project_name` in Prometheus metrics. It can be used for filtering in Grafana dashboard.