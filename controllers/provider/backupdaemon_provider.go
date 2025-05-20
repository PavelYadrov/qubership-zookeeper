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
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

type BackupDaemonResourceProvider struct {
	cr          *zookeeperservice.ZooKeeperService
	logger      logr.Logger
	spec        *zookeeperservice.BackupDaemon
	serviceName string
}

func NewBackupDaemonResourceProvider(cr *zookeeperservice.ZooKeeperService, logger logr.Logger) BackupDaemonResourceProvider {
	return BackupDaemonResourceProvider{
		cr:          cr,
		logger:      logger,
		spec:        cr.Spec.BackupDaemon,
		serviceName: fmt.Sprintf("%s-backup-daemon", cr.Name),
	}
}

func (bdrp BackupDaemonResourceProvider) GetServiceName() string {
	return bdrp.serviceName
}

// NewBackupDaemonClientService returns a client service for ZooKeeper Backup Daemon
func (bdrp BackupDaemonResourceProvider) NewBackupDaemonClientService() *corev1.Service {
	backupDaemonLabels := bdrp.GetBackupDaemonLabels()
	selectorLabels := bdrp.GetBackupDaemonSelectorLabels()
	backupDaemonPort := bdrp.getBackupDaemonPort()
	ports := []corev1.ServicePort{
		{
			Name:     "http",
			Port:     backupDaemonPort,
			Protocol: corev1.ProtocolTCP,
		},
	}
	return newServiceForCR(bdrp.serviceName, bdrp.cr.Namespace, backupDaemonLabels, selectorLabels, ports)
}

// NewBackupDaemonDeployment returns a deployment for ZooKeeper Backup Daemon
func (bdrp BackupDaemonResourceProvider) NewBackupDaemonDeployment() *appsv1.Deployment {
	backupDaemonLabels := bdrp.GetBackupDaemonLabels()
	backupDaemonLabels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", bdrp.serviceName, bdrp.cr.Namespace)
	backupDaemonLabels["app.kubernetes.io/technology"] = "python"
	selectorLabels := bdrp.GetBackupDaemonSelectorLabels()
	backupDaemonCustomLabels := bdrp.GetBackupDaemonCustomLabels(backupDaemonLabels)
	replicas := int32(1)
	volumes := bdrp.getBackupDaemonVolumes()
	volumeMounts := bdrp.getBackupDaemonVolumeMounts()
	backupDaemonPort := bdrp.getBackupDaemonPort()
	ports := []corev1.ContainerPort{
		{
			ContainerPort: backupDaemonPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	envVars := []corev1.EnvVar{
		{
			Name:  "SERVICE_NAME",
			Value: bdrp.serviceName,
		},
		{
			Name:  "ZOOKEEPER_HOST",
			Value: bdrp.spec.ZooKeeperHost,
		},
		{
			Name:  "ZOOKEEPER_PORT",
			Value: strconv.Itoa(bdrp.spec.ZooKeeperPort),
		},
		{
			Name:  "PV_TYPE",
			Value: bdrp.spec.BackupStorage.PersistentVolumeType,
		},
		{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "ZOOKEEPER_ENABLE_SSL",
			Value: strconv.FormatBool(bdrp.cr.Spec.Global.ZooKeeperSsl.Enabled),
		},
	}

	if bdrp.spec.BackupSchedule != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BACKUP_SCHEDULE",
			Value: bdrp.spec.BackupSchedule,
		})
	}
	if bdrp.spec.EvictionPolicy != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "EVICTION_POLICY",
			Value: bdrp.spec.EvictionPolicy,
		})
	}
	if bdrp.spec.IPv6 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BROADCAST_ADDRESS",
			Value: "::",
		})
	}

	if bdrp.spec.S3 != nil && bdrp.spec.S3.Enabled {
		s3Envs := []corev1.EnvVar{
			{
				Name:  "S3_ENABLED",
				Value: strconv.FormatBool(bdrp.spec.S3.Enabled),
			},
			{
				Name:  "S3_URL",
				Value: bdrp.spec.S3.Url,
			},
			{
				Name:  "S3_BUCKET",
				Value: bdrp.spec.S3.Bucket,
			},
		}
		envVars = append(envVars, s3Envs...)
	}

	envVars = append(envVars, bdrp.getZooKeeperCredentialsEnvs()...)

	if IsVaultSecretManagementEnabled(bdrp.cr) {
		envVars = append(envVars, getVaultConnectionEnvVars(bdrp.GetServiceName(), bdrp.cr)...)
		volumes = append(volumes, corev1.Volume{Name: "vault-env", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "vault-env", MountPath: "/vault"})
	}

	if bdrp.cr.Spec.Global.ZooKeeperSsl.Enabled && bdrp.cr.Spec.Global.ZooKeeperSsl.SecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "ssl-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: bdrp.cr.Spec.Global.ZooKeeperSsl.SecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "ssl-certs", MountPath: "/tls"})
	}

	if bdrp.cr.Spec.BackupDaemon.BackupDaemonSsl.Enabled && bdrp.cr.Spec.BackupDaemon.BackupDaemonSsl.SecretName != "" {
		envVars = append(envVars, []corev1.EnvVar{
			{Name: "TLS_ENABLED", Value: "true"},
			{Name: "CERTS_PATH", Value: "/backupTLS"},
		}...)
		volumes = append(volumes, corev1.Volume{
			Name: "backup-ssl-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: bdrp.cr.Spec.BackupDaemon.BackupDaemonSsl.SecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "backup-ssl-certs", MountPath: "/backupTLS"})
	}

	if bdrp.spec.S3 != nil && bdrp.spec.S3.Enabled && bdrp.spec.S3.SslVerify && bdrp.spec.S3.SslCert != "" {
		envVars = append(envVars, []corev1.EnvVar{
			{Name: "S3_CERTS_PATH", Value: "/s3Certs"},
		}...)
		volumes = append(volumes, corev1.Volume{
			Name: "s3-ssl-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: bdrp.spec.S3.SslSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "s3-ssl-certs", MountPath: "/s3Certs"})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bdrp.serviceName,
			Namespace: bdrp.cr.Namespace,
			Labels:    backupDaemonLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: backupDaemonCustomLabels,
				},
				Spec: corev1.PodSpec{
					Volumes:        volumes,
					InitContainers: bdrp.getInitContainers(),
					Containers: []corev1.Container{
						{
							Name:            bdrp.serviceName,
							Image:           bdrp.spec.DockerImage,
							Ports:           ports,
							Env:             envVars,
							Resources:       bdrp.spec.Resources,
							VolumeMounts:    volumeMounts,
							LivenessProbe:   bdrp.getLivenessProbe(),
							ReadinessProbe:  bdrp.getReadinessProbe(),
							ImagePullPolicy: corev1.PullAlways,
							Command:         bdrp.getCommand(),
							Args:            bdrp.getArgs(),
							SecurityContext: getDefaultContainerSecurityContext(),
						},
					},
					SecurityContext:    &bdrp.spec.SecurityContext,
					ServiceAccountName: bdrp.GetServiceAccountName(),
					Hostname:           bdrp.serviceName,
					Affinity:           bdrp.getBackupDaemonAffinityRules(),
					Tolerations:        bdrp.spec.Tolerations,
					PriorityClassName:  bdrp.spec.PriorityClassName,
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
		},
	}
}

// getBackupDaemonVolumes configures the list of ZooKeeper Backup Daemon volumes
func (bdrp BackupDaemonResourceProvider) getBackupDaemonVolumes() []corev1.Volume {
	var volumeSource corev1.VolumeSource
	if bdrp.spec.BackupStorage.PersistentVolumeType == "" {
		volumeSource = corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}
	} else {
		backupPersistentVolumeClaimName := bdrp.spec.BackupStorage.PersistentVolumeClaimName
		if backupPersistentVolumeClaimName == "" {
			backupPersistentVolumeClaimName = fmt.Sprintf(SnapshotsPersistentVolumeClaimPattern, bdrp.cr.Name)
		}
		volumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: backupPersistentVolumeClaimName,
			},
		}
	}
	return []corev1.Volume{
		{
			Name:         "backup-storage",
			VolumeSource: volumeSource,
		},
	}
}

// getBackupDaemonVolumeMounts configures the list of ZooKeeper Backup Daemon volume mounts
func (bdrp BackupDaemonResourceProvider) getBackupDaemonVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "backup-storage",
			ReadOnly:  false,
			MountPath: "/opt/zookeeper/backup-storage",
		},
	}
}

// getLivenessProbe configures the liveness probe for ZooKeeper Backup Daemon
func (bdrp BackupDaemonResourceProvider) getLivenessProbe() *corev1.Probe {
	probe := bdrp.getProbe()
	backupDaemonPort := int(bdrp.getBackupDaemonPort())
	probe.Handler = corev1.Handler{
		TCPSocket: &corev1.TCPSocketAction{
			Port: intstr.FromInt(backupDaemonPort),
		},
	}
	return probe
}

// getReadinessProbe configures the readiness probe for ZooKeeper Backup Daemon
func (bdrp BackupDaemonResourceProvider) getReadinessProbe() *corev1.Probe {
	return bdrp.getLivenessProbe()
}

// getProbe configures common parameters for liveness and readiness probe for ZooKeeper Backup Daemon
func (bdrp BackupDaemonResourceProvider) getProbe() *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    5,
	}
}

func (bdrp BackupDaemonResourceProvider) getBackupDaemonAffinityRules() *corev1.Affinity {
	affinityRules := bdrp.spec.Affinity.DeepCopy()
	if bdrp.spec.BackupStorage.NodeName != "" {
		affinityRules.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{bdrp.spec.BackupStorage.NodeName},
							},
						},
					},
				},
			},
		}
	}
	return affinityRules
}

func (bdrp BackupDaemonResourceProvider) getCommand() []string {
	if IsVaultSecretManagementEnabled(bdrp.cr) {
		return []string{"/vault/vault-env"}
	}
	return nil
}

func (bdrp BackupDaemonResourceProvider) getArgs() []string {
	if IsVaultSecretManagementEnabled(bdrp.cr) {
		return []string{"python3", "/opt/backup/backup-daemon.py"}
	}
	return nil
}

func (bdrp BackupDaemonResourceProvider) getInitContainers() []corev1.Container {
	if IsVaultSecretManagementEnabled(bdrp.cr) {
		return []corev1.Container{
			getVaultInitContainer(bdrp.cr),
		}
	}
	return nil
}

func (bdrp BackupDaemonResourceProvider) getZooKeeperCredentialsEnvs() []corev1.EnvVar {
	var envs []corev1.EnvVar
	if IsVaultSecretManagementEnabled(bdrp.cr) {
		envs = []corev1.EnvVar{
			{
				Name:  "BACKUP_DAEMON_API_CREDENTIALS_USERNAME",
				Value: getVaultSecretEnvVarSource(bdrp.serviceName, bdrp.cr, "credentials", "username"),
			},
			{
				Name:  "BACKUP_DAEMON_API_CREDENTIALS_PASSWORD",
				Value: getVaultSecretEnvVarSource(bdrp.serviceName, bdrp.cr, "credentials", "password"),
			},
			{
				Name:  "ZOOKEEPER_ADMIN_USERNAME",
				Value: getVaultSecretEnvVarSource(bdrp.cr.Name, bdrp.cr, "admin-credentials", "username"),
			},
			{
				Name:  "ZOOKEEPER_ADMIN_PASSWORD",
				Value: getVaultSecretEnvVarSource(bdrp.cr.Name, bdrp.cr, "admin-credentials", "password"),
			},
		}
	} else {
		envs = []corev1.EnvVar{
			{
				Name:      "BACKUP_DAEMON_API_CREDENTIALS_USERNAME",
				ValueFrom: getSecretEnvVarSource(bdrp.spec.SecretName, "username"),
			},
			{
				Name:      "BACKUP_DAEMON_API_CREDENTIALS_PASSWORD",
				ValueFrom: getSecretEnvVarSource(bdrp.spec.SecretName, "password"),
			},
			{
				Name:      "ZOOKEEPER_ADMIN_USERNAME",
				ValueFrom: getSecretEnvVarSource(bdrp.spec.SecretName, "zookeeper-admin-username"),
			},
			{
				Name:      "ZOOKEEPER_ADMIN_PASSWORD",
				ValueFrom: getSecretEnvVarSource(bdrp.spec.SecretName, "zookeeper-admin-password"),
			},
		}
	}

	if bdrp.spec.S3 != nil && bdrp.spec.S3.Enabled {
		s3Envs := []corev1.EnvVar{
			{
				Name:      "S3_KEY_ID",
				ValueFrom: getSecretEnvVarSource(bdrp.spec.S3.SecretName, "s3-key-id"),
			},
			{
				Name:      "S3_KEY_SECRET",
				ValueFrom: getSecretEnvVarSource(bdrp.spec.S3.SecretName, "s3-key-secret"),
			},
		}
		envs = append(envs, s3Envs...)
	}
	return envs
}

// GetBackupDaemonLabels configures common labels for ZooKeeper Backup Daemon resources
func (bdrp BackupDaemonResourceProvider) GetBackupDaemonLabels() map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = bdrp.serviceName
	labels = util.JoinMaps(util.JoinMaps(labels, bdrp.GetBackupDaemonSelectorLabels()), bdrp.cr.Spec.Global.DefaultLabels)
	return labels
}

func (bdrp BackupDaemonResourceProvider) GetBackupDaemonSelectorLabels() map[string]string {
	return map[string]string{
		"name":      bdrp.serviceName,
		"component": "zookeeper-backup-daemon",
	}
}

func (bdrp BackupDaemonResourceProvider) GetBackupDaemonCustomLabels(backupDaemonLabels map[string]string) map[string]string {
	globalLabels := bdrp.cr.Spec.Global.CustomLabels
	customLabels := bdrp.spec.CustomLabels
	return util.JoinMaps(util.JoinMaps(globalLabels, customLabels), backupDaemonLabels)
}

// GetServiceAccountName returns service account name for pods. Now it's equal to service name.
func (bdrp BackupDaemonResourceProvider) GetServiceAccountName() string {
	return bdrp.GetServiceName()
}

func (bdrp BackupDaemonResourceProvider) getBackupDaemonPort() int32 {
	if bdrp.spec.BackupDaemonSsl.Enabled {
		return 8443
	} else {
		return 8080
	}
}
