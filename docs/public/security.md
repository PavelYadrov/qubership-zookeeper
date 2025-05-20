This section provides information about exposed ports, user accounts, password policies and password changing procedures and ZooKeeper security measures.

## Exposed Ports

List of ports used by ZooKeeper and other Services.

| Port | Service                     | Description                                                                                                                                                                                                                                                                                |
|------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 2181 | ZooKeeper                   | Port of ZooKeeper.                                                                                                                                                                                                                                                                         |
| 2182 | ZooKeeper                   | Port used for non-encrypted access to ZooKeeper.                                                                                                                                                                                                                                           |
| 2888 | Zookeeper                   | Port	used for communication between follower nodes and the leader for syncing data.                                                                                                                                                                                                         |
| 3888 | Zookeeper                   | Port	used for leader election among nodes in the zooKeeper ensemble.                                                                                                                                                                                                                        |
| 8080 | Zookeeper                   | Port used for monitoring prometheus.                                                                                                                                                                                                                                                       |
| 8081 | Zookeeper                   | Port	used for backup and restoration tasks related to zooKeeper.                                                                                                                                                                                                                            |
| 8096 | Zookeeper                   | Port used for prometheus monitoring if `monitoring.monitoringType` parameter value is equal to `prometheus`.                                                                                                                                                                               |
| 9087 | Zookeeper                   | Port used for jolokia agent. This agent defines the JMX-HTTP bridge.                                                                                                                                                                                                                       |
| 8080 | Zookeeper Integration-tests | Exposes the container's port to the network. It allows access to the application running in the container.                                                                                                                                                                                 |
| 8125 | Zookeeper Monitoring        | Port used for StatsD metrics collection.                                                                                                                                                                                                                                                   |
| 8094 | Zookeeper Monitoring        | Port used for collecting or transmitting monitoring data over TCP.                                                                                                                                                                                                                         |
| 8092 | Zookeeper Monitoring        | Port used for collecting or transmitting monitoring data over UDP.                                                                                                                                                                                                                         |
| 8443 | Backup Daemon               | Port is used for secure communication with the backup daemon service when TLS is enabled. This ensures encrypted and secure data transmission.                                                                                                                                             |
| 8080 | Backup Daemon               | Port used to manage and execute backup and restoration tasks to ensure data integrity and availability.                                                                                                                                                                                    |
| 9443 | Controller Manager          | Port on which the controller's webhook server is listening.                                                                                                                                                                                                                                |
| 8081 | Controller Manager          | Port is used for both liveness and readiness probes of the controller-manager container. The liveness probe checks the /healthz endpoint to ensure the container is alive, while the readiness probe checks the /readyz endpoint to determine if the container is ready to serve requests. |

## User Accounts

List of user accounts used for Zookeeper and other Services.

| Service                 | OOB accounts | Deployment parameter                     | Is Break Glass account | Can be blocked | Can be deleted | Comment                                                                                                                                        |
|-------------------------|--------------|------------------------------------------|------------------------|----------------|----------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Zookeeper               | admin        | global.secrets.zooKeeper.adminUsername   | yes                    | no             | no             | The default admin user. There is no default value, the name can be specified during deploy. Otherwise ZooKeeper is deployed without security.  |
| Zookeeper               | client       | global.secrets.zooKeeper.clientUsername  | no                     | yes            | yes            | The default client user. There is no default value, the name can be specified during deploy. Otherwise ZooKeeper is deployed without security. |
| ZooKeeper Backup Daemon | client       | global.secrets.backupDaemon.username     | no                     | yes            | yes            | ZooKeeper Backup Daemon user. There is no default value, the name must be specified during deploy.                                             |
| Zookeeper               | client       | global.secrets.zooKeeper.additionalUsers | no                     | yes            | yes            | Additional zookeeper users. There is no default value, the name must be specified during deploy.                                               |

## Disabling User Accounts

Zookeeper does not support disabling user accounts.

## Password Policies

* Passwords must be at least 8 characters long. This ensures a basic level of complexity and security.
* The passwords can contain only the following symbols:
    * Alphabets: a-zA-Z
    * Numerals: 0-9
    * Punctuation marks: ., ;, !, ?
    * Mathematical symbols: -, +, *, /, %
    * Brackets: (, ), {, }, <, >
    * Additional symbols: _, |, &, @, $, ^, #, ~

**Note**: To ensure that passwords are sufficiently complex, it is recommended to include:

* A minimum length of 8 characters
* At least one uppercase letter (A-Z)
* At least one lowercase letter (a-z)
* At least one numeral (0-9)
* At least one special character from the allowed symbols list

## Changing password guide

Password changing procedures for ZooKeeper Service is described in respective guide:

* [Password changing guide](/docs/public/password-changing.md)

# Authentication

ZooKeeper allows protecting connections by authentication (SASL: Digest-MD5), but it does not restrict other clients from connecting without authentication.
Therefore, authentication is not useful without authorization via ACL.
By default, SASL authentication is disabled, and can be enabled by specifying the `global.secrets.zooKeeper.adminUsername` and `global.secrets.zooKeeper.adminPassword` properties during installation.

## ZooKeeper Server Security Properties

* The `global.secrets.zooKeeper.adminUsername` parameter specifies the username of the ZooKeeper administrator user. These credentials are used by ZooKeeper nodes to communicate.
  With `global.secrets.zooKeeper.adminPassword`, it enables ZooKeeper authentication, and can be specified explicitly during installation.
* The `global.secrets.zooKeeper.adminPassword` parameter specifies the password of the ZooKeeper administrator user. These credentials are used by ZooKeeper nodes to communicate.
  With `global.secrets.zooKeeper.adminUsername`, it enables ZooKeeper authentication, and can be specified explicitly during installation.
* The `global.secrets.zooKeeper.clientUsername` parameter specifies the username of the ZooKeeper client user.
  These credentials are used by the ZooKeeper client to establish connection with the ZooKeeper server.
* The `global.secrets.zooKeeper.clientPassword` parameter specifies the password of the ZooKeeper client user.
  These credentials are used by the ZooKeeper client to establish connection with the ZooKeeper server.
* The `global.secrets.zooKeeper.additionalUsers` parameter specifies comma-separated pairs (`username:password`) of additional users that are used by clients for authentication
  in ZooKeeper. For example, `user1_name:user1_password,user2_name:user2_password`.
* The `zooKeeper.quorumAuthEnabled` parameter enables internal authentication between ZooKeeper nodes.

## ZooKeeper Clients Security Properties

To authenticate, clients must send credentials within a connection request. It requires the following settings:

1. Create the JAAS configuration file:

    ```text
    cat >> ${ZOOKEEPER_HOME}/conf/client_jaas.conf << EOL
    Client {
               org.apache.zookeeper.server.auth.DigestLoginModule required
               username="username"
               password="password";
        };
    EOL
    ```

    The `username` and `password` properties are used by the clients to initiate connections to the ZooKeeper server.

2. Specify path to the JAAS configuration file in the JVM property `java.security.auth.login.config`.

    For example, for `zkCli.sh` client, it is necessary to use the following command:

    ```sh
    export CLIENT_JVMFLAGS="-Djava.security.auth.login.config=${ZOOKEEPER_HOME}/conf/client_jaas.conf" && \
    ./bin/zkCli.sh
    ```

# Authorization

ZooKeeper supports pluggable authentication schemes. It uses Access Control Lists (ACLs) to control access to its znodes (data nodes of a ZooKeeper data tree).

The ACL implementation is quite similar to UNIX file access permissions. A node may have any number of `<scheme:expression,perms>` pairs.
The left member of the pair specifies the authentication scheme, while the right member indicates permissions (ACL pertains only to a specific znode).
For more information about ZooKeeper ACL, refer to _ZooKeeper Programmer's Guide_ [https://zookeeper.apache.org/doc/r3.5.8/zookeeperProgrammers.html#sc_ZooKeeperAccessControl](https://zookeeper.apache.org/doc/r3.5.5/zookeeperProgrammers.html#sc_ZooKeeperAccessControl).

SASL ACL looks like `sasl:{username}:{set of permissions}`.

ZooKeeper supports the following permissions:

* `CREATE` - Create a child node.
* `READ` - Get data from a node and list its children.
* `WRITE` - Set data for a node.
* `DELETE` - Delete a child node.
* `ADMIN` - Set permissions.

The following features of ZooKeeper ACL are very important:

* The absence of `DELETE` permissions does not restrict delete operations for ZooKeeper parent node, only for child nodes. But you can forbid deletion if there is a child node under the parent node.

    The `CREATE` and `DELETE` permissions have been broken out of the `WRITE` permission for finer grained access controls.

    The cases for `CREATE` and `DELETE` permissions are the following:

    * `WRITE` without `CREATE` and `DELETE` - You are able to do a `set` operation on a ZooKeeper node, but not able to create or delete children.

    * `CREATE` without `DELETE` - Clients create requests by creating ZooKeeper nodes in a parent directory. You want all clients to be able to add, but only the request processor can delete.

* ZooKeeper ACLs do not propagate hierarchically.
  For example, if you want to restrict `WRITE` permission for all child nodes of some parent ZooKeeper node, you have to set necessary ACL for all these nodes.

It has the following built-in schemes:

* `world` – Anyone.
* `sasl` – For kerberos/sasl authentication.
* `digest` – For MD5 hash.
* `ip` – IP used as ACL ID identity.

Using `zkCli.sh` client, you can set ACL with two options:

* When creating znode:

  ```sh
  create /test data world:anyone:r, sasl:client:crdwa
  ```

  This ACL allows all users (including anonymous) to read node `/test`, but restricts other permissions and allows the SASL user `client` all permissions for this node.

* Using `setAcl` command:

  ```sh
  setAcl /test sasl:client:crdw, sasl:admin:crdwa
  ```

  This ACL allows the SASL user `client` the permissions `create`, `read`, `delete`, and `write`, but restricts `admin` permission and allows the SASL user `admin` all permissions.
  For other users, all permissions are restricted.

# Logging

Security events and critical operations should be logged for audit purposes. You can find more details about enabling
audit logging in [Audit Guide](/docs/public/audit.md).
