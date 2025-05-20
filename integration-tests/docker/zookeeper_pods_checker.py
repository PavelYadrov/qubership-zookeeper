import os
import time

from PlatformLibrary import PlatformLibrary

environ = os.environ
namespace = environ.get("ZOOKEEPER_OS_PROJECT")
service = environ.get("ZOOKEEPER_HOST")
backup_daemon = environ.get("ZOOKEEPER_BACKUP_DAEMON_HOST")
managed_by_operator = environ.get("ZOOKEEPER_IS_MANAGED_BY_OPERATOR")
timeout = 300


if __name__ == '__main__':
    time.sleep(10)
    try:
        k8s_lib = PlatformLibrary(managed_by_operator)
    except Exception as e:
        print(e)
        exit(1)
    timeout_start = time.time()
    while time.time() < timeout_start + timeout:
        try:
            deployments = k8s_lib.get_deployment_entities_count_for_service(namespace, service)
            ready_deployments = k8s_lib.get_active_deployment_entities_count_for_service(namespace, service)
            if backup_daemon is not None and len(backup_daemon) != 0:
                deployments += k8s_lib.get_deployment_entities_count_for_service(namespace, backup_daemon, 'component')
                ready_deployments += k8s_lib.get_active_deployment_entities_count_for_service(namespace, backup_daemon, 'component')
            print(f'[Check status] deployments: {deployments}, ready deployments: {ready_deployments}')
        except Exception as e:
            print(e)
            continue
        if deployments == ready_deployments and deployments != 0:
            print("ZooKeeper deployments are ready")
            exit(0)
        time.sleep(10)
    print(f'ZooKeeper deployments are not ready at least {timeout} seconds')
    exit(1)