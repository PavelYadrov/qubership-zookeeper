// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"fmt"
	zookeeperservice "github.com/Netcracker/qubership-zookeeper/api/v1"
	"github.com/Netcracker/qubership-zookeeper/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

const (
	persistentVolumeClaimPattern = "pvc-%s-%d"
	devMode                      = "dev"
	prodMode                     = "prod"
)

type ZooKeeperResourceProvider struct {
	cr     *zookeeperservice.ZooKeeperService
	logger logr.Logger
	spec   *zookeeperservice.ZooKeeper
}

func NewZooKeeperResourceProvider(cr *zookeeperservice.ZooKeeperService, logger logr.Logger) ZooKeeperResourceProvider {
	return ZooKeeperResourceProvider{
		cr:     cr,
		logger: logger,
		spec:   cr.Spec.ZooKeeper,
	}
}

func (zrp ZooKeeperResourceProvider) GetServiceName() string {
	return zrp.cr.Name
}

// NewZooKeeperClientServiceForCR returns the client service for ZooKeeper
func (zrp ZooKeeperResourceProvider) NewZooKeeperClientServiceForCR() *corev1.Service {
	zooKeeperLabels := GetZooKeeperLabels(zrp.cr.Name, zrp.cr.Spec.Global.DefaultLabels)
	selectorLabels := GetZooKeeperSelectorLabels(zrp.cr.Name)
	ports := []corev1.ServicePort{
		{Name: "zookeeper-client", Port: 2181, Protocol: corev1.ProtocolTCP},
		{Name: "nonencrypted-zookeeper-client", Port: 2182, Protocol: corev1.ProtocolTCP},
	}
	clientService := newServiceForCR(zrp.cr.Name, zrp.cr.Namespace, zooKeeperLabels, selectorLabels, ports)
	return clientService
}

// NewZooKeeperDomainServiceForCR returns the domain service for ZooKeeper
func (zrp ZooKeeperResourceProvider) NewZooKeeperDomainServiceForCR() *corev1.Service {
	serviceName := fmt.Sprintf("%s-server", zrp.cr.Name)
	zooKeeperLabels := GetZooKeeperLabels(zrp.cr.Name, zrp.cr.Spec.Global.DefaultLabels)
	zooKeeperLabels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", "zookeper", zrp.cr.Namespace)
	selectorLabels := GetZooKeeperSelectorLabels(zrp.cr.Name)
	ports := []corev1.ServicePort{
		{Name: "zookeeper-client", Port: 2181, Protocol: corev1.ProtocolTCP},
		{Name: "nonencrypted-zookeeper-client", Port: 2182, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-followers", Port: 2888, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-election", Port: 3888, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-backup", Port: 8081, Protocol: corev1.ProtocolTCP},
	}
	domainService := newServiceForCR(serviceName, zrp.cr.Namespace, zooKeeperLabels, selectorLabels, ports)
	domainService.Spec.ClusterIP = "None"
	domainService.Spec.PublishNotReadyAddresses = true
	return domainService
}

// NewZooKeeperServerServiceForCR returns a service for specified ZooKeeper server
func (zrp ZooKeeperResourceProvider) NewZooKeeperServerServiceForCR(serverId int) *corev1.Service {
	serviceName := fmt.Sprintf("%s-%d", zrp.cr.Name, serverId)
	zooKeeperLabels := GetZooKeeperLabels(zrp.cr.Name, zrp.cr.Spec.Global.DefaultLabels)
	zooKeeperLabels["name"] = serviceName
	selectorLabels := GetZooKeeperSelectorLabels(zrp.cr.Name)
	selectorLabels["name"] = serviceName
	ports := []corev1.ServicePort{
		{Name: "zookeeper-client", Port: 2181, Protocol: corev1.ProtocolTCP},
		{Name: "nonencrypted-zookeeper-client", Port: 2182, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-followers", Port: 2888, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-election", Port: 3888, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-backup", Port: 8081, Protocol: corev1.ProtocolTCP},
		{Name: "zookeeper-jolokia", Port: zrp.spec.JolokiaPort, Protocol: corev1.ProtocolTCP},
		{Name: "prometheus-http", Port: 8080, Protocol: corev1.ProtocolTCP},
	}
	serverService := newServiceForCR(serviceName, zrp.cr.Namespace, zooKeeperLabels, selectorLabels, ports)
	return serverService
}

// NewZooKeeperPersistentVolumeClaimForCR returns a persistent volume claim for specified ZooKeeper server
func (zrp ZooKeeperResourceProvider) NewZooKeeperPersistentVolumeClaimForCR(serverId int) *corev1.PersistentVolumeClaim {
	var persistentVolumeName string
	if len(zrp.spec.Storage.Volumes) > 0 {
		persistentVolumeName = zrp.spec.Storage.Volumes[serverId-1]
	}
	var persistentVolumeLabel string
	if len(zrp.spec.Storage.Labels) > 0 {
		persistentVolumeLabel = zrp.spec.Storage.Labels[serverId-1]
	}
	var storageClassName *string
	if len(zrp.spec.Storage.ClassName) > 0 {
		storageClassName = &zrp.spec.Storage.ClassName[0]
		if len(zrp.spec.Storage.ClassName) == zrp.spec.Replicas {
			storageClassName = &zrp.spec.Storage.ClassName[serverId-1]
		}
	}
	persistentVolumeClaimName := fmt.Sprintf(persistentVolumeClaimPattern, zrp.cr.Name, serverId)
	return ProcessNonSharedPersistentVolumeClaim(persistentVolumeClaimName, persistentVolumeName, persistentVolumeLabel,
		storageClassName, zrp.spec.Storage.Size, zrp.cr.Namespace, GetZooKeeperLabels(zrp.cr.Name, zrp.cr.Spec.Global.DefaultLabels), zrp.logger)
}

// NewServerDeploymentForCR returns a deployment for specified ZooKeeper server
func (zrp ZooKeeperResourceProvider) NewServerDeploymentForCR(serverId int) *appsv1.Deployment {
	deploymentName := fmt.Sprintf("%s-%d", zrp.cr.Name, serverId)
	domainName := fmt.Sprintf("%s-server", zrp.cr.Name)
	zooKeeperLabels := GetZooKeeperLabels(zrp.cr.Name, zrp.cr.Spec.Global.DefaultLabels)
	zooKeeperLabels["name"] = deploymentName
	zooKeeperLabels["app.kubernetes.io/technology"] = "java-others"
	zooKeeperLabels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", zrp.cr.Name, zrp.cr.Namespace)
	selectorLabels := GetZooKeeperSelectorLabels(zrp.cr.Name)
	selectorLabels["name"] = deploymentName
	zooKeeperCustomLabels := zrp.GetZooKeeperCustomLabels(zooKeeperLabels)
	replicas := int32(1)
	livenessProbe := corev1.Probe{
		Handler: corev1.Handler{
			Exec: zrp.getExecCommand([]string{"./bin/zkHealth.sh", "liveness-probe"}),
		},
		InitialDelaySeconds: 20,
		TimeoutSeconds:      20,
		PeriodSeconds:       20,
		SuccessThreshold:    1,
		FailureThreshold:    5,
	}
	readinessProbe := corev1.Probe{
		Handler: corev1.Handler{
			Exec: zrp.getExecCommand([]string{"./bin/zkHealth.sh", "readiness-probe"}),
		},
		InitialDelaySeconds: 40,
		TimeoutSeconds:      20,
		PeriodSeconds:       20,
		SuccessThreshold:    1,
		FailureThreshold:    5,
	}
	var dataVolumeSource corev1.VolumeSource
	if len(zrp.spec.Storage.Volumes) > 0 || len(zrp.spec.Storage.Labels) > 0 || len(zrp.spec.Storage.ClassName) > 0 {
		dataVolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf(persistentVolumeClaimPattern, zrp.cr.Name, serverId),
			},
		}
	} else {
		dataVolumeSource = corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}
	}
	var backupVolumeSource corev1.VolumeSource
	if zrp.spec.SnapshotStorage.PersistentVolumeType == "" || zrp.spec.SnapshotStorage.PersistentVolumeType == "standalone" {
		backupVolumeSource = corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}
	} else {
		snapshotsPersistentVolumeClaimName := zrp.spec.SnapshotStorage.PersistentVolumeClaimName
		if snapshotsPersistentVolumeClaimName == "" {
			snapshotsPersistentVolumeClaimName = fmt.Sprintf(SnapshotsPersistentVolumeClaimPattern, zrp.cr.Name)
		}
		backupVolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: snapshotsPersistentVolumeClaimName,
			},
		}
	}
	envVars := []corev1.EnvVar{
		{Name: "SERVER_NAME", Value: zrp.cr.Name},
		{Name: "SERVER_ID", Value: strconv.Itoa(serverId)},
		{Name: "SERVER_DOMAIN", Value: domainName},
		{
			Name: "SERVER_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{Name: "SERVER_COUNT", Value: strconv.Itoa(zrp.spec.Replicas)},
		{
			Name:  "HEAP_OPTS",
			Value: fmt.Sprintf("-Xms%dm -Xmx%dm", zrp.spec.HeapSize, zrp.spec.HeapSize),
		},
		{
			Name:  "QUORUM_AUTH_ENABLED",
			Value: strconv.FormatBool(zrp.spec.QuorumAuthEnabled),
		},
		{
			Name:  "JOLOKIA_PORT",
			Value: strconv.Itoa(int(zrp.spec.JolokiaPort)),
		},
		{
			Name:  "AUDIT_ENABLED",
			Value: strconv.FormatBool(zrp.spec.AuditEnabled),
		},
	}

	envVars = append(envVars, zrp.getSecretEnvs()...)

	volumes := []corev1.Volume{
		{Name: "data", VolumeSource: dataVolumeSource},
		{Name: "log", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		{Name: "backup-storage", VolumeSource: backupVolumeSource},
	}

	volumeMounts := []corev1.VolumeMount{
		{Name: "data", MountPath: "/var/opt/zookeeper/data"},
		{Name: "log", MountPath: "/opt/zookeeper/log"},
		{Name: "backup-storage", MountPath: "/opt/zookeeper/backup-storage"},
	}

	diagnosticMode := zrp.spec.Diagnostics.Mode
	if diagnosticMode == devMode || diagnosticMode == prodMode {
		envVars = append(envVars, []corev1.EnvVar{
			{Name: "CLOUD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
			{Name: "MICROSERVICE_NAME", Value: zrp.GetServiceName()},
			{Name: "NC_DIAGNOSTIC_MODE", Value: diagnosticMode},
			{Name: "NC_DIAGNOSTIC_AGENT_SERVICE", Value: zrp.spec.Diagnostics.AgentService},
		}...)
	}

	if IsVaultSecretManagementEnabled(zrp.cr) {
		envVars = append(envVars, getVaultConnectionEnvVars(zrp.GetServiceName(), zrp.cr)...)
		volumes = append(volumes, corev1.Volume{Name: "vault-env", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "vault-env", MountPath: "/vault"})
	}

	if zrp.cr.Spec.Global.ZooKeeperSsl.Enabled && zrp.cr.Spec.Global.ZooKeeperSsl.SecretName != "" {
		envVars = append(envVars, []corev1.EnvVar{
			{Name: "ENABLE_SSL", Value: "true"},
			{Name: "SSL_CIPHER_SUITES", Value: strings.Join(zrp.cr.Spec.ZooKeeper.Ssl.CipherSuites, ",")},
			{Name: "ENABLE_2WAY_SSL", Value: strconv.FormatBool(zrp.cr.Spec.ZooKeeper.Ssl.EnableTwoWaySsl)},
			{Name: "ALLOW_NONENCRYPTED_ACCESS", Value: strconv.FormatBool(zrp.cr.Spec.ZooKeeper.Ssl.AllowNonencryptedAccess)},
		}...)

		volumes = append(volumes, corev1.Volume{
			Name: "ssl-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: zrp.cr.Spec.Global.ZooKeeperSsl.SecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "ssl-certs", MountPath: "/opt/zookeeper/tls"})
	}

	serverDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: zrp.cr.Namespace,
			Labels:    zooKeeperLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: zooKeeperCustomLabels},
				Spec: corev1.PodSpec{
					Volumes:        volumes,
					InitContainers: zrp.getInitContainers(),
					Containers: []corev1.Container{
						{
							Name:    "zookeeper",
							Command: zrp.getCommand(),
							Args:    zrp.getArgs(),
							Image:   zrp.spec.DockerImage,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 2181, Protocol: corev1.ProtocolTCP},
								{ContainerPort: 2182, Protocol: corev1.ProtocolTCP},
								{ContainerPort: 2888, Protocol: corev1.ProtocolTCP},
								{ContainerPort: 3888, Protocol: corev1.ProtocolTCP},
								{ContainerPort: zrp.spec.JolokiaPort, Protocol: corev1.ProtocolTCP},
								{ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
							},
							LivenessProbe:   &livenessProbe,
							ReadinessProbe:  &readinessProbe,
							Env:             buildEnvs(envVars, zrp.spec.EnvironmentVariables, zrp.logger),
							Resources:       zrp.spec.Resources,
							VolumeMounts:    volumeMounts,
							ImagePullPolicy: corev1.PullAlways,
							SecurityContext: getDefaultContainerSecurityContext(),
						},
					},
					SecurityContext:    &zrp.spec.SecurityContext,
					Hostname:           deploymentName,
					ServiceAccountName: zrp.GetServiceAccountName(),
					Subdomain:          domainName,
					Affinity:           zrp.getZooKeeperAffinityRules(serverId),
					Tolerations:        zrp.spec.Tolerations,
					PriorityClassName:  zrp.spec.PriorityClassName,
				},
			},
		},
	}
	return serverDeployment
}

func (zrp ZooKeeperResourceProvider) GetZooKeeperCustomLabels(zooKeeperLabels map[string]string) map[string]string {
	globalLabels := zrp.cr.Spec.Global.CustomLabels
	customLabels := zrp.spec.CustomLabels
	return util.JoinMaps(util.JoinMaps(globalLabels, customLabels), zooKeeperLabels)
}

func (zrp ZooKeeperResourceProvider) getCommand() []string {
	if IsVaultSecretManagementEnabled(zrp.cr) {
		return []string{"/vault/vault-env"}
	}
	return nil
}

func (zrp ZooKeeperResourceProvider) getArgs() []string {
	if IsVaultSecretManagementEnabled(zrp.cr) {
		return []string{"/sbin/tini", "--", "/docker-entrypoint.sh", "start"}
	}
	return nil
}

func (zrp ZooKeeperResourceProvider) getInitContainers() []corev1.Container {
	if IsVaultSecretManagementEnabled(zrp.cr) {
		return []corev1.Container{
			getVaultInitContainer(zrp.cr),
		}
	}
	return nil
}

func (zrp ZooKeeperResourceProvider) getExecCommand(originalCommand []string) *corev1.ExecAction {
	if IsVaultSecretManagementEnabled(zrp.cr) {
		originalCommand = append([]string{"/vault/vault-env"}, originalCommand...)
	}
	return &corev1.ExecAction{Command: originalCommand}
}

func (zrp ZooKeeperResourceProvider) getSecretEnvs() []corev1.EnvVar {
	if IsVaultSecretManagementEnabled(zrp.cr) {
		return []corev1.EnvVar{
			{
				Name:  "ADMIN_USERNAME",
				Value: getVaultSecretEnvVarSource(zrp.GetServiceName(), zrp.cr, "admin-credentials", "username"),
			},
			{
				Name:  "ADMIN_PASSWORD",
				Value: getVaultSecretEnvVarSource(zrp.GetServiceName(), zrp.cr, "admin-credentials", "password"),
			},
			{
				Name:  "CLIENT_USERNAME",
				Value: getVaultSecretEnvVarSource(zrp.GetServiceName(), zrp.cr, "client-credentials", "username"),
			},
			{
				Name:  "CLIENT_PASSWORD",
				Value: getVaultSecretEnvVarSource(zrp.GetServiceName(), zrp.cr, "client-credentials", "password"),
			},
			{
				Name:  "ADDITIONAL_USERS",
				Value: getVaultSecretEnvVarSource(zrp.GetServiceName(), zrp.cr, "additional-users", "users"),
			},
		}
	} else {
		return []corev1.EnvVar{
			{
				Name:      "ADMIN_USERNAME",
				ValueFrom: getSecretEnvVarSource(zrp.spec.SecretName, "admin-username"),
			},
			{
				Name:      "ADMIN_PASSWORD",
				ValueFrom: getSecretEnvVarSource(zrp.spec.SecretName, "admin-password"),
			},
			{
				Name:      "CLIENT_USERNAME",
				ValueFrom: getSecretEnvVarSource(zrp.spec.SecretName, "client-username"),
			},
			{
				Name:      "CLIENT_PASSWORD",
				ValueFrom: getSecretEnvVarSource(zrp.spec.SecretName, "client-password"),
			},
			{
				Name:      "ADDITIONAL_USERS",
				ValueFrom: getSecretEnvVarSource(zrp.spec.SecretName, "additional-users"),
			},
		}
	}
}

// getZooKeeperAffinityRules configures the ZooKeeper affinity rules
func (zrp ZooKeeperResourceProvider) getZooKeeperAffinityRules(serverId int) *corev1.Affinity {
	affinityRules := zrp.spec.Affinity.DeepCopy()
	if len(zrp.spec.Storage.Nodes) > 0 {
		affinityRules.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{zrp.spec.Storage.Nodes[serverId-1]},
							},
						},
					},
				},
			},
		}
	}
	return affinityRules
}

// GetServiceAccountName returns service account name for pods. Now it's equal to service name.
func (zrp ZooKeeperResourceProvider) GetServiceAccountName() string {
	return zrp.GetServiceName()
}
