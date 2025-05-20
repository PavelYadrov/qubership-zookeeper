This section provides information on the password changing procedures in the ZooKeeper service.

# ZooKeeper

This section provides information about changing password for ZooKeeper server.

## ZooKeeper Server Credentials

The ZooKeeper secret contains user credentials for SASL authentication in ZooKeeper server.
During deployment, if no credentials are specified, ZooKeeper is deployed with SASL authentication disabled.

To update ZooKeeper credentials:

1. Navigate to **OpenShift > ${PROJECT_NAME} > Resources > Secrets**.
2. Select the secret name **${SERVICE_NAME}-secret**.
3. Navigate to **Actions** > **Edit YAML**.
4. Update the values for the `admin-username`, `admin-password`, `client-username`, `client-password`, and `additional-users` properties with new credentials in `BASE64` encoding.
   For more information about these credentials, refer to [ZooKeeper Security](/docs/public/security.md) section in the _Zookeeper Service Installation Procedure_.
5. Click **Save**.
6. Restart ZooKeeper to apply the newly specified credentials.

where:

* `${PROJECT_NAME}` is the name of the OpenShift project where ZooKeeper is.
* `${SERVICE_NAME}` is the name of the ZooKeeper service.

**Note:** If you change the credentials for the ZooKeeper server, you also have to change the credentials for client services (Kafka, [ZooKeeper Backup Daemon](#zookeeper-backup-daemon)).
For more information about the password changing procedure, refer to the _Cloud Platform Maintenance Guide_.

**Important:** If you want to update the ZooKeeper secret in a DR scheme, it is necessary to perform all steps for the left (`left-${SERVICE_NAME}`) and right (`right-${SERVICE_NAME}`) sides.

# ZooKeeper Backup Daemon

This section provides information about changing password for ZooKeeper backup daemon.

## ZooKeeper Credentials

The ZooKeeper Backup Daemon credentials secret contains the admin user credentials to connect to ZooKeeper.
For correct functioning, ZooKeeper Backup Daemon credentials must already be defined in [ZooKeeper Server Credentials](#zookeeper-server-credentials). Update these credentials accordingly.

To update the ZooKeeper Backup Daemon credentials secret:

1. Navigate to **OpenShift > ${PROJECT_NAME} > Resources > Secrets**.
1. Select the secret name **${SERVICE_NAME}-secret**.
1. Navigate to **Actions** > **Edit YAML**.
1. Update the values of the `zookeeper-admin-username` and `zookeeper-admin-password` properties with new credentials in `BASE64` encoding.
1. Click **Save**.
1. Restart ZooKeeper Backup Daemon to apply the newly specified credentials.

where:

* `${PROJECT_NAME}` is the name of the OpenShift project where ZooKeeper Backup Daemon is.
* `${SERVICE_NAME}` is the name of the ZooKeeper Backup Daemon service.

**Important:** If you want to update the ZooKeeper Backup Daemon secret in a DR scheme, it is necessary to perform all steps for the left (`left-${SERVICE_NAME}`) and right (`right-${SERVICE_NAME}`) sides.
