{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "zookeeper-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "zookeeper-service.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "zookeeper-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "zookeeper-service.globalPodSecurityContext" -}}
runAsNonRoot: true
seccompProfile:
  type: "RuntimeDefault"
{{- with .Values.global.securityContext }}
{{ toYaml . }}
{{- end -}}
{{- end -}}

{{- define "zookeeper-service.globalContainerSecurityContext" -}}
allowPrivilegeEscalation: false
capabilities:
  drop: ["ALL"]
{{- end -}}

{{/*
Common labels
*/}}
{{- define "zookeeper-service.labels" -}}
helm.sh/chart: {{ include "zookeeper-service.chart" . }}
{{ include "zookeeper-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "zookeeper-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "zookeeper-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "zookeeper-service.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "zookeeper-service.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Whether ZooKeeper SSL enabled
*/}}
{{- define "zookeeper-service.enableSsl" -}}
    {{- and .Values.global.tls.enabled .Values.zooKeeper.tls.enabled -}}
{{- end -}}

{{/*
Whether BackupDaemon SSL enabled
*/}}
{{- define "backupDaemon.enableSsl" -}}
  {{- and .Values.global.tls.enabled .Values.backupDaemon.install .Values.backupDaemon.tls.enabled -}}
{{- end -}}


{{/*
Whether ZooKeeper certificates are specified
*/}}
{{- define "zookeeper.certificatesSpecified" -}}
  {{- $filled := false -}}
  {{- range $key, $value := .Values.zooKeeper.tls.certificates -}}
    {{- if $value -}}
        {{- $filled = true -}}
    {{- end -}}
  {{- end -}}
  {{- $filled -}}
{{ end }}

{{/*
ZooKeeper SSL secret name
*/}}
{{- define "zookeeper-service.sslSecretName" -}}
  {{- if and .Values.global.tls.enabled .Values.zooKeeper.tls.enabled -}}
    {{- if and (or .Values.global.tls.generateCerts.enabled (eq (include "zookeeper.certificatesSpecified" .) "true")) (not .Values.zooKeeper.tls.secretName) -}}
      {{- printf "%s-tls-secret" (include "zookeeper.name" .) -}}
    {{- else -}}
      {{- .Values.zooKeeper.tls.secretName -}}
    {{- end -}}
  {{- else -}}
    {{- "" -}}
  {{- end -}}
{{- end -}}

{{/*
Vault enabled
*/}}
{{- define "vault.enabled" -}}
{{- if .Values.vaultSecretManagement }}{{- if .Values.vaultSecretManagement.enabled }}
    {{- printf "true" }}
{{- else }}
    {{- printf "false" }}
{{- end }}
{{- else }}
    {{- printf "false" }}
{{- end -}}
{{- end -}}

{{/*
Find a vault-env image in various places.
Image can be found from:
* SaaS/App deployer (or groovy.deploy.v3) from .Values.deployDescriptor "vault-env" "image"
* DP.Deployer from .Values.deployDescriptor.vaultEnv.image
* or from default values .Values.vaultSecretManagement.dockerImage
*/}}
{{- define "vaultEnv.image" -}}
    {{- printf "%s" .Values.vaultSecretManagement.dockerImage -}}
{{- end -}}

{{/*
ZooKeeper service name.
*/}}
{{- define "zookeeper.name" -}}
{{- coalesce .Values.global.name .Values.name "zookeeper" -}}
{{- end -}}

{{/*
Compute the minimum number of available ZooKeeper replicas for the PodDisruptionBudget.
This defaults to (n/2)+1 where n is the number of members of the server cluster.
*/}}
{{- define "zookeeper.pdb.minAvailable" -}}
{{- if le (int (include "zookeeper.replicas" . )) 1 -}}
{{ 0 }}
{{- else if .Values.zooKeeper.disruptionBudget.minAvailable -}}
{{ .Values.zooKeeper.disruptionBudget.minAvailable -}}
{{- else -}}
{{- add (div (int (include "zookeeper.replicas" . )) 2) 1 -}}
{{- end -}}
{{- end -}}

{{/*
DNS names used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "zookeeper.certDnsNames" -}}
  {{- $zookeeperName := include "zookeeper.name" . -}}
  {{- $dnsNames := list "localhost" $zookeeperName (printf "%s.%s" $zookeeperName .Release.Namespace) (printf "%s.%s" $zookeeperName "zookeeper-server") (printf "%s.%s.%s" $zookeeperName "zookeeper-server" .Release.Namespace) (printf "%s.%s.svc" $zookeeperName .Release.Namespace) -}}
  {{- $brokers := include "zookeeper.replicas" . -}}
  {{- $zookeeperNamespace := .Release.Namespace -}}
  {{- range $i, $e := until ($brokers | int) -}}
    {{- $dnsNames = append $dnsNames (printf "%s-%d" $zookeeperName (add $i 1)) -}}
    {{- $dnsNames = append $dnsNames (printf "%s-%d.%s" $zookeeperName (add $i 1) $zookeeperNamespace) -}}
    {{- $dnsNames = append $dnsNames (printf "%s-%d.zookeeper-server.%s" $zookeeperName (add $i 1) $zookeeperNamespace) -}}
  {{- end -}}
  {{- $dnsNames = concat $dnsNames .Values.zooKeeper.tls.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}

{{/*
IP addresses used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "zookeeper.certIpAddresses" -}}
  {{- $ipAddresses := list "127.0.0.1" -}}
  {{- $ipAddresses = concat $ipAddresses .Values.zooKeeper.tls.subjectAlternativeName.additionalIpAddresses -}}
  {{- $ipAddresses | toYaml -}}
{{- end -}}

{{/*
Generate certificates for ZooKeeper server
*/}}
{{- define "zookeeper.generateCerts" -}}
  {{- $dnsNames := include "zookeeper.certDnsNames" . | fromYamlArray -}}
  {{- $ipAddresses := include "zookeeper.certIpAddresses" . | fromYamlArray -}}
  {{- $duration := default 365 .Values.global.tls.generateCerts.durationDays | int -}}
  {{- $ca := genCA "zookeeper-ca" $duration -}}
  {{- $zookeeperName := include "zookeeper.name" . -}}
  {{- $cert := genSignedCert $zookeeperName $ipAddresses $dnsNames $duration $ca -}}
tls.crt: {{ $cert.Cert | b64enc }}
tls.key: {{ $cert.Key | b64enc }}
ca.crt: {{ $ca.Cert | b64enc }}
{{- end -}}

{{/*
Provider used to generate SSL certificates
*/}}
{{- define "services.certProvider" -}}
  {{- default "helm" .Values.global.tls.generateCerts.certProvider -}}
{{- end -}}


{{/*
Find a zookeeper-integration-tests image in various places.
*/}}
{{- define "zookeeper-integration-tests.image" -}}
    {{- printf "%s" .Values.integrationTests.image -}}
{{- end -}}

{{/*
Find a zookeeper-service-operator image in various places.
*/}}
{{- define "zookeeper-service-operator.image" -}}
    {{- printf "%s" .Values.operator.dockerImage -}}
{{- end -}}

{{/*
Find a zookeeper image in various places.
*/}}
{{- define "zookeeper.image" -}}
    {{- printf "%s" .Values.zooKeeper.dockerImage -}}
{{- end -}}

{{/*
ZooKeeper admin username.
*/}}
{{- define "zookeeper.adminUsername" -}}
  {{- if and (ne (.Values.INFRA_ZOOKEEPER_ADMIN_USERNAME | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.INFRA_ZOOKEEPER_ADMIN_USERNAME }}
  {{- else -}}
    {{- .Values.global.secrets.zooKeeper.adminUsername -}}
  {{- end -}}
{{- end -}}

{{/*
ZooKeeper admin password.
*/}}
{{- define "zookeeper.adminPassword" -}}
  {{- if and (ne (.Values.INFRA_ZOOKEEPER_ADMIN_PASSWORD | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.INFRA_ZOOKEEPER_ADMIN_PASSWORD }}
  {{- else -}}
    {{- .Values.global.secrets.zooKeeper.adminPassword -}}
  {{- end -}}
{{- end -}}

{{/*
ZooKeeper client username.
*/}}
{{- define "zookeeper.clientUsername" -}}
  {{- if and (ne (.Values.INFRA_ZOOKEEPER_CLIENT_USERNAME | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.INFRA_ZOOKEEPER_CLIENT_USERNAME }}
  {{- else -}}
    {{- .Values.global.secrets.zooKeeper.clientUsername -}}
  {{- end -}}
{{- end -}}

{{/*
ZooKeeper client password.
*/}}
{{- define "zookeeper.clientPassword" -}}
  {{- if and (ne (.Values.INFRA_ZOOKEEPER_CLIENT_PASSWORD | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.INFRA_ZOOKEEPER_CLIENT_PASSWORD }}
  {{- else -}}
    {{- .Values.global.secrets.zooKeeper.clientPassword -}}
  {{- end -}}
{{- end -}}

{{/*
ZooKeeper replicas.
*/}}
{{- define "zookeeper.replicas" -}}
  {{- if and (ne (.Values.INFRA_ZOOKEEPER_REPLICAS | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.INFRA_ZOOKEEPER_REPLICAS }}
  {{- else -}}
    {{- default 3 .Values.zooKeeper.replicas -}}
  {{- end -}}
{{- end -}}

{{/*
Storage class from various places.
*/}}
{{- define "zookeeper.storageClassName" -}}
  {{- if and (ne (.Values.STORAGE_RWO_CLASS | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.STORAGE_RWO_CLASS | toStrings }}
  {{- else -}}
    {{- default "" .Values.zooKeeper.storage.className -}}
  {{- end -}}
{{- end -}}

{{/*
Find a zookeeper-monitoring image in various places.
*/}}
{{- define "zookeeper-monitoring.image" -}}
    {{- printf "%s" .Values.monitoring.dockerImage -}}
{{- end -}}

{{/*
ZooKeeper Monitoring installation required
*/}}
{{- define "monitoring.install" -}}
  {{- if and (ne (.Values.MONITORING_ENABLED | toString) "<nil>") .Values.global.cloudIntegrationEnabled -}}
    {{- .Values.MONITORING_ENABLED }}
  {{- else -}}
    {{- .Values.monitoring.install -}}
  {{- end -}}
{{- end -}}

{{/*
Find a zookeeper-backup-daemon image in various places.
*/}}
{{- define "zookeeper-backup-daemon.image" -}}
    {{- printf "%s" .Values.backupDaemon.dockerImage -}}
{{- end -}}

{{- define "monitoring.zookeeperHost" -}}
  {{- $zookeeperHosts := list -}}
  {{- range $replica, $e := until ((include "zookeeper.replicas" . ) | int) -}}
    {{- $zookeeperHosts = append $zookeeperHosts (printf "'%s-%d:2181'" (include "zookeeper.name" $) (add1 $replica)) }}
  {{- end -}}
  {{- join "," $zookeeperHosts -}}
{{- end -}}

{{/*
Backup Daemon Protocol
*/}}
{{- define "backupDaemon.Protocol" -}}
  {{- if and .Values.global.tls.enabled .Values.backupDaemon.tls.enabled -}}
    {{- "https" -}}
  {{- else -}}
    {{- "http" -}}
  {{- end -}}
{{- end -}}

{{/*
Whether Backup Daemon certificates are Specified
*/}}
{{- define "backupDaemon.certificatesSpecified" -}}
  {{- $filled := false -}}
  {{- range $key, $value := .Values.backupDaemon.tls.certificates -}}
    {{- if $value -}}
        {{- $filled = true -}}
    {{- end -}}
  {{- end -}}
  {{- $filled -}}
{{ end }}

{{/*
Backup Daemon SSL secret name
*/}}
{{- define "backupDaemon.sslSecretName" -}}
  {{- if and .Values.global.tls.enabled .Values.backupDaemon.install .Values.backupDaemon.tls.enabled -}}
    {{- if and (or .Values.global.tls.generateCerts.enabled (eq (include "backupDaemon.certificatesSpecified" .) "true")) (not .Values.backupDaemon.tls.secretName) -}}
      {{- printf "%s-backup-daemon-tls-secret" (include "zookeeper.name" .) -}}
    {{- else -}}
      {{- .Values.backupDaemon.tls.secretName -}}
    {{- end -}}
  {{- else -}}
    {{- "" -}}
  {{- end -}}
{{- end -}}

{{/*
Backup Daemon S3 SSL secret name
*/}}
{{- define "backupDaemon.s3.tlsSecretName" -}}
  {{- if .Values.backupDaemon.s3.sslCert -}}
    {{- if .Values.backupDaemon.s3.sslSecretName -}}
      {{- .Values.backupDaemon.s3.sslSecretName -}}
    {{- else -}}
      {{- printf "zookeeper-backup-daemon-s3-tls-secret" -}}
    {{- end -}}
  {{- else -}}
    {{- if .Values.backupDaemon.s3.sslSecretName -}}
      {{- .Values.backupDaemon.s3.sslSecretName -}}
    {{- else -}}
      {{- printf "" -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
DNS names used to generate SSL certificate with "Subject Alternative Name" field for Backup Daemon
*/}}
{{- define "backupDaemon.certDnsNames" -}}
  {{- $zookeeperName := include "zookeeper.name" . -}}
  {{- $backupDaemonNamespace := .Release.Namespace -}}
  {{- $dnsNames := list "localhost" (printf "%s-backup-daemon" $zookeeperName) (printf "%s-backup-daemon.%s" $zookeeperName $backupDaemonNamespace) (printf "%s-backup-daemon.%s.svc.cluster.local" $zookeeperName $backupDaemonNamespace) -}}
  {{- $dnsNames = concat $dnsNames .Values.backupDaemon.tls.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}

{{/*
IP addresses used to generate SSL certificate with "Subject Alternative Name" field for Backup Daemon
*/}}
{{- define "backupDaemon.certIpAddresses" -}}
  {{- $ipAddresses := list "127.0.0.1" -}}
  {{- $ipAddresses = concat $ipAddresses .Values.backupDaemon.tls.subjectAlternativeName.additionalIpAddresses -}}
  {{- $ipAddresses | toYaml -}}
{{- end -}}

{{/*
Generate certificates for Backup Daemon
*/}}
{{- define "backupDaemon.generateCerts" -}}
  {{- $dnsNames := include "backupDaemon.certDnsNames" . | fromYamlArray -}}
  {{- $ipAddresses := include "backupDaemon.certIpAddresses" . | fromYamlArray -}}
  {{- $duration := default 365 .Values.global.tls.generateCerts.durationDays | int -}}
  {{- $ca := genCA "zookeeper-backup-daemon-ca" $duration -}}
  {{- $backupDaemonName := "backupDaemon" -}}
  {{- $cert := genSignedCert $backupDaemonName $ipAddresses $dnsNames $duration $ca -}}
tls.crt: {{ $cert.Cert | b64enc }}
tls.key: {{ $cert.Key | b64enc }}
ca.crt: {{ $ca.Cert | b64enc }}
{{- end -}}


{{/*
TLS Static Metric secret template
Arguments:
Dictionary with:
* "namespace" is a namespace of application
* "application" is name of application
* "service" is a name of service
* "enabledSsl" is ssl enabled for service
* "secret" is a name of tls secret for service
* "certProvider" is a type of tls certificates provider
* "certificate" is a name of CertManger's Certificate resource for service
Usage example:
{{template "global.tlsStaticMetric" (dict "namespace" .Release.Namespace "application" .Chart.Name "service" .global.name "enabledSsl" (include "global.sslEnabled" .) "secret" (include "global.sslSecretName" .) "certProvider" (include "services.certProvider" .) "certificate" (printf "%s-tls-certificate" (include "global.name")) }}
*/}}
{{- define "global.tlsStaticMetric" -}}
- expr: {{ ternary "1" "0" (eq .enabledSsl "true") }}
  labels:
    namespace: "{{ .namespace }}"
    application: "{{ .application }}"
    service: "{{ .service }}"
    {{ if eq .enabledSsl "true" }}
    secret: "{{ .secret }}"
    {{ if eq .certProvider "cert-manager" }}
    certificate: "{{ .certificate }}"
    {{ end }}
    {{ end }}
  record: service:tls_status:info
{{- end -}}

{{ define "zookeeper-service.findImage" }}
  {{- $root := index . 0 -}}
  {{- $service_name := index . 1 -}}
  {{- if index $root.Values.deployDescriptor $service_name }}
  {{- index $root.Values.deployDescriptor $service_name "image" }}
  {{- else }}
  {{- "not_found" }}
  {{- end }}
{{- end }}

{{- define "zookeeper-service.monitoredImages" -}}
  {{- printf "deployment %s-service-operator zookeeper-service-operator %s, " (include "zookeeper.name" .) (include "zookeeper-service.findImage" (list . "zookeeper-service-operator")) -}}
  {{- if gt (int (include "zookeeper.replicas" .)) 0 -}}
    {{- printf "deployment %s-1 zookeeper %s, " (include "zookeeper.name" .) (include "zookeeper-service.findImage" (list . "docker-zookeeper")) -}}
  {{- end -}}
  {{- if (eq (include "monitoring.install" .) "true") }}
    {{- printf "deployment %s-monitoring zookeeper-monitoring %s, " (include "zookeeper.name" .) (include "zookeeper-service.findImage" (list . "zookeeper-monitoring")) -}}
  {{- end -}}
  {{- if .Values.backupDaemon.install }}
    {{- printf "deployment %s-backup-daemon zookeeper-backup-daemon %s, " (include "zookeeper.name" .) (include "zookeeper-service.findImage" (list . "zookeeper-backup-daemon")) -}}
  {{- end -}}
  {{- if .Values.integrationTests.install }}
    {{- printf "deployment %s-integration-tests-runner zookeeper-integration-tests-runner %s, " (include "zookeeper.name" .) (include "zookeeper-service.findImage" (list . "zookeeper-integration-tests")) -}}
  {{- end -}}
{{- end }}

{{/*
Common Zookeeper chart related resources labels
*/}}
{{- define "zookeeper.defaultLabels" -}}
app.kubernetes.io/version: '{{ .Values.ARTIFACT_DESCRIPTOR_VERSION | trunc 63 | trimAll "-_." }}'
app.kubernetes.io/component: 'backend'
app.kubernetes.io/part-of: '{{ .Values.PART_OF }}'
{{- end -}}