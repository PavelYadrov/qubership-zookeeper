ZooKeeper service allows storing credentials in Vault secrets.
If you enable this option, the ZooKeeper operator generates passwords for user names specified during deployment and stores them to the Vault secrets.

ZooKeeper pods in turn use the service account token to login to the Vault and obtain values of secrets before running.

The following is an example of a common case deployment parameters for enabled Vault secret management:

```yaml
vaultSecretManagement:
  enabled: true
  url: http://vault-service-vault.example.kubernetes.com
  method: kubernetes
  role: kubernetes-operator-role
  path: secret
  writePolicies: true
  passwordGenerationMechanism: operator
  refreshCredentials: false
```

# Vault Prerequisites

Before deploying ZooKeeper, ensure that Vault is working correctly and ready to accept client requests.
For more information and commands regarding Vault configuration, refer to the [Official Vault Documentation](https://www.vaultproject.io/docs).

To allow Operator and ZooKeeper service store credentials in Vault it is necessary to prepare the following configurations on the Vault side:

* Configure the Authentication method for Kubernetes/OpenShift. The name of configured method is used as value for the `vaultSecretManagement.method` deployment parameter.
* Create the Secret storage with version 2. The name of secret storage is used as value for `vaultSecretManagement.path` deployment parameter.
  By default, the Vault is deployed with secret storage `secret` which can be used to store ZooKeeper credentials.
* Create the Role for Kafka operator with the corresponding rights. The name of role is used as value for the `vaultSecretManagement.role` deployment parameter.
  This role should be assigned for service account name: `zookeeper-service-operator`.

For more information about role with admin rights, see [Deploy with Predefined Operator Admin Rights](#deploy-with-predefined-operator-admin-rights) and for role without admin rights,
see [Deploy without Predefined Operator Admin Rights](#deploy-without-predefined-operator-admin-rights).

## Deploy with Predefined Operator Admin Rights

If the Vault provides the operators rights for policy creation, the ZooKeeper Operator can create corresponding policies and roles for ZooKeeper, ZooKeeper Monitoring and ZooKeeper Backup Daemon automatically.
The property `vaultSecretManagement.writePolicies` should be set to `true`.

The Vault is deployed with default policy `operator-policy` and role `kubernetes-operator-role` which can be used for ZooKeeper deployment.
If there are not operator policies or roles, you need to create them.

Ensure that the default Vault operator policy contains the following rights and you need to add if any rights are missing:

```hcl
path "sys/policies/*" {
  capabilities = ["create", "read", "update","delete", "list"]
}
path "/auth/kubernetes/role/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "/secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

Where `secret` is the name of secret storage and it should be equal to the `path` deployment parameter.

Ensure that the default Vault operator role parameters  `bound_service_account_names` and `bound_service_account_namespaces` contain name and namespace of ZooKeeper Operator service account.
For example, `zookeeper-service-operator` and `zookeeper-service`. You need to add it if they are missing.

## Deploy without Predefined Operator Admin Rights

If the Vault does not provide the operators rights for policy creation then you need to create the required policies manually.

The Policies and Roles described in the sections below contain corresponding configurations and you must create them under Vault token with rights before ZooKeeper deployment.

For all the configurations in the sections below:

* `secret` is the name of secret storage and it should be equal to the `path` deployment parameter.
* `zookeeper` and `zookeeper-service` are ZooKeeper service and namespace names.

### ZooKeeper Operator

**Policy**

zookeeper-service-operator.zookeeper-service-policy:

```hcl
path "/secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

**Role**

zookeeper-service-operator.zookeeper-service-role:

```text
bound_service_account_names=zookeeper-service-operator
bound_service_account_namespaces=zookeeper-service
policies=zookeeper-service-operator.zookeeper-service
```

### ZooKeeper

**Admin Policy**

zookeeper.zookeeper-service-admin-policy:

```hcl
path "secret/data/zookeeper.zookeeper-service/*" 
{
	capabilities = ["read", "list"]
}
```

**Client Policy**

zookeeper.zookeeper-service-client-policy:

```hcl
path "secret/data/zookeeper.zookeeper-service/client-credentials" 
{
	capabilities = ["read", "list"]
}
```

**Role**

zookeeper.zookeeper-service-role:

```text
bound_service_account_names=zookeeper
bound_service_account_namespaces=zookeeper-service
policies=zookeeper.zookeeper-service-admin-policy
```

### ZooKeeper Monitoring

**Policy**

zookeeper-monitoring.zookeeper-service-policy:

```hcl
path "secret/data/zookeeper-monitoring.zookeeper-service/*" 
{
	capabilities = ["read", "list"]
}
```

**Role**

zookeeper-monitoring.zookeeper-service-role:

```text
bound_service_account_names=zookeeper-monitoring
bound_service_account_namespaces=zookeeper-service
policies=zookeeper-monitoring.zookeeper-service-policy,zookeeper.zookeeper-service-client-policy
```

### ZooKeeper Backup Daemon

**Policy**

zookeeper-backup-daemon.zookeeper-service-policy:

```hcl
path "secret/data/zookeeper-backup-daemon.zookeeper-service/*" 
{
    capabilities = ["read", "list"]
}
```

**Role**

zookeeper-backup-daemon.zookeeper-service-role:

```text
bound_service_account_names=zookeeper-backup-daemon
bound_service_account_namespaces=zookeeper-service
policies=zookeeper-backup-daemon.zookeeper-service-policy,zookeeper.zookeeper-service-admin-policy
```

# Credentials Rotation

To refresh ZooKeeper credentials it is necessary to perform the following steps:

1. Perform `upgrade` job for ZooKeeper with previous parameters and set value of the `vaultSecretManagement.refreshCredentials` parameter to `true`.
   After this, the Operator generates new passwords for all ZooKeeper secrets.
2. Restart all ZooKeeper services (ZooKeeper, ZooKeeper Monitoring, and ZooKeeper Backup Daemon) and all services which use Vault ZooKeeper secrets to connect. For example, Kafka.
