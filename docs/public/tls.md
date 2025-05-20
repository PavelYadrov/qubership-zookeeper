# TLS

## SSL Configuration using predefined secret with TLS certificates

You can enable TLS-based encryption for communication with ZooKeeper. For this you need to do the following:

1. Create secret with certificates (or specify to generate them automatically using `generateCerts` deployment parameter):

    ```yaml
    kind: Secret
    apiVersion: v1
    metadata:
      name: ${SECRET_NAME}
      namespace: ${NAMESPACE}
    data:
      ca.crt: ${ROOT_CA_CERTIFICATE}
      tls.crt: ${CERTIFICATE}
      tls.key: ${PRIVATE_KEY}
    type: kubernetes.io/tls
    ```

    Where:
    * `${SECRET_NAME}` is the name of secret that contains all certificates. For example, `zookeeper-tls-secret`.
    * `${NAMESPACE}` is the namespace where the secret should be created. For example, `zookeeper-service`.
    * `${ROOT_CA_CERTIFICATE}` is the root CA in BASE64 format.
    * `${CERTIFICATE}` is the certificate in BASE64 format.
    * `${PRIVATE_KEY}` is the private key in BASE64 format.

2. Specify the following deployment parameters:

    ```yaml
    zooKeeper:
      ...
      ssl:
        enabled: true
        secretName: "zookeeper-tls-secret"
        cipherSuites: []
        allowNonencryptedAccess: false
        generateCerts:
          enabled: false
          certProvider: helm
          duration: 365
          clusterIssuerName: ""
          subjectAlternativeName:
            additionalDnsNames: []
            additionalIpAddresses: []
    ```

## SSL Configuration using cert-manager

The example of parameters to deploy ZooKeeper with enabled TLS and `CertManager` certificate generation:

```yaml
zooKeeper:
  ...
  ssl:
    enabled: true
    generateCerts:
      enabled: true
      certProvider: cert-manager
      duration: 365
      clusterIssuerName: "Cluster Issuer Name"
```

## SSL Configuration using helm

You can automatically generate TLS-based secrets using Helm by specifying certificates in deployment parameters. For example, to generate `zookeeper-tls-secret`:

1. Prepare certificates in BASE64 encoded format:

    ```yaml
    ca.crt: ${ROOT_CA_CERTIFICATE}
    tls.crt: ${CERTIFICATE}
    tls.key: ${PRIVATE_KEY}
    ```

    Where:

      * ${ROOT_CA_CERTIFICATE} is the root CA in BASE64 format.
      * ${CERTIFICATE} is the certificate in BASE64 format.
      * ${PRIVATE_KEY} is the private key in BASE64 format.

2. Specify TLS settings and certificates in Deployer job parameters.
This setup is designed for deploying ZooKeeper with TLS enabled, utilizing pre-generated certificates.
The `zooKeeper.tls.secretName` parameter allows you to specify a custom Kubernetes Secret name.
If it is not provided, a default name (`<global.name>-tls-secret`) based on your deployment will be used.

The example of parameters to deploy Zookeeper service with enabled TLS and Helm certificate provider:

  ```yaml
    global:
      name: zookeeper
      waitForPodsReady: true
      podReadinessTimeout: 250
      tls:
        enabled: true
        cipherSuites: []
        allowNonencryptedAccess: false
        generateCerts:
          enabled: false
          certProvider: helm
          durationDays: 365
          clusterIssuerName: ""
    zooKeeper:
      storage:
        className:
          - local-path
        size: 2Gi
      heapSize: 255
      tls:
      enabled: true
      certificates:
        crt: LS0tLS1CRUdJTiBDRVJU......
        key: LS0tLS1CRUdJTiBSU0EgUFJ....
        ca: LS0tLS1CRUdtuguyhuhkjij.....
    backupDaemon:
      install: true
      tls:
        enabled: true
        certificates:
          crt: LS0tLS1CRUdJTiBDRVJU......
          key: LS0tLS1CRUdJTiBSU.....
          ca: LS0tLS1CRUdtuguyhuhkjij....
  ```

**NOTE:** Use only `Rolling Update` mode because `Clean Install` mode deletes all pre-created secrets.

## Certificate Renewal

CertManager automatically renews Certificates.
It calculates when to renew a Certificate based on the issued X.509 certificate's duration and a `renewBefore` value
which specifies how long before expiry a certificate should be renewed.
By default, the value of `renewBefore` parameter is 2/3 through the X.509 certificate's `duration`.
More info in [Cert Manager Renewal](https://cert-manager.io/docs/usage/certificate/#renewal).

After certificate renewed by CertManager the secret contains new certificate, but running applications store previous
certificate in pods.
As CertManager generates new certificates before old expired the both certificates are valid for some time (`renewBefore`).

ZooKeeper service does not have any handlers for certificates secret changes, so you need to manually restart **all**
ZooKeeper service pods until the time when old certificate is expired.
