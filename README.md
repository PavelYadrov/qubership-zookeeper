# ZooKeeper Service

## Repository structure

* `./demo` - directory `docker-compose` to run ZooKeeper, integration tests & backup-daemon.
* `./docs` - directory with actual documentation for users and developers.
* `./charts/helm/zookeeper-service` - directory with **main** HELM chart with resources for ZooKeeper and
  integration tests.
* `./integration-tests` - directory with Robot Framework test cases for ZooKeeper.

## How to start

### Deploy to k8s

#### Pure helm

1. Build operator and integration tests, if you need non-master versions.
2. Prepare kubeconfig on you host machine to work with target cluster.
3. Prepare `sample.yaml` file with deployment parameters, which should contains custom docker images if it is needed.
4. Store `sample.yaml` file in `charts/helm/zookeeper-service` directory.
5. Go to `charts/helm/zookeeper-service` directory.
6. Run the following command:

  ```sh
  helm install zookeeper-service ./ -f sample.yaml -n <TARGET_NAMESPACE>
  ```

### Smoke tests

There is no smoke tests.

### How to debug

#### VSCode

To debug zookeeper-operator in VSCode you can use `Launch operator` configuration which is already defined in 
`./.vscode/launch.json` file.

The developer should configure environment variables: `WATCH_NAMESPACE`, `KUBECONFIG`.

Regarding `KUBECONFIG`, developer should **need to define** `KUBECONFIG` environment variable
which should contains path to the kube-config file. It can be defined on configuration level
or on the level of user's environment variables.

### How to troubleshoot

There are no well-defined rules for troubleshooting, as each task is unique, but there are some tips that can do:

* Deploy parameters.
* Application manifest.
* Logs from all ZooKeeper service pods: operator, ZooKeeper, monitoring, backup-daemon.

Also, developer can take a look on [Troubleshooting guide](/docs/public/troubleshooting.md).

## Evergreen strategy

To keep the component up to date, the following activities should be performed regularly:

* Vulnerabilities fixing.
* ZooKeeper upgrade.
* Bug-fixing, improvement and feature implementation for operator and other related supplementary services.

## Useful links

* [ZooKeeper Quickstart guide](/docs/internal/quickstart.md).
* [Installation guide](/docs/public/installation.md).
* [Troubleshooting guide](/docs/public/troubleshooting.md).
* [Internal Developer Guide](/docs/internal/developing.md).