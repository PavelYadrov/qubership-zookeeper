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
)

type MonitoringResourceProvider struct {
	cr          *zookeeperservice.ZooKeeperService
	logger      logr.Logger
	spec        *zookeeperservice.Monitoring
	serviceName string
}

func NewMonitoringResourceProvider(cr *zookeeperservice.ZooKeeperService, logger logr.Logger) MonitoringResourceProvider {
	return MonitoringResourceProvider{
		cr:          cr,
		logger:      logger,
		spec:        cr.Spec.Monitoring,
		serviceName: fmt.Sprintf("%s-monitoring", cr.Name),
	}
}

func (mrp MonitoringResourceProvider) GetServiceName() string {
	return mrp.serviceName
}

// NewMonitoringClientService returns a client service for ZooKeeper Monitoring
func (mrp MonitoringResourceProvider) NewMonitoringClientService() *corev1.Service {
	monitoringLabels := mrp.GetMonitoringLabels()
	selectorLabels := mrp.GetMonitoringSelectorLabels()
	ports := []corev1.ServicePort{
		{
			Name:     "zookeeper-monitoring-statsd",
			Port:     8125,
			Protocol: corev1.ProtocolTCP,
		},
		{
			Name:     "zookeeper-monitoring-tcp",
			Port:     8094,
			Protocol: corev1.ProtocolTCP,
		},
		{
			Name:     "zookeeper-monitoring-udp",
			Port:     8092,
			Protocol: corev1.ProtocolUDP,
		},
	}
	if mrp.spec.MonitoringType == "prometheus" {
		ports = append(ports, corev1.ServicePort{
			Name:     "prometheus-cli",
			Port:     8096,
			Protocol: corev1.ProtocolTCP,
		})
	}
	clientService := newServiceForCR(mrp.serviceName, mrp.cr.Namespace, monitoringLabels, selectorLabels, ports)
	return clientService
}

// NewMonitoringDeployment returns a deployment for ZooKeeper Monitoring
func (mrp MonitoringResourceProvider) NewMonitoringDeployment() *appsv1.Deployment {
	monitoringLabels := mrp.GetMonitoringLabels()
	monitoringLabels["app.kubernetes.io/technology"] = "python"
	monitoringLabels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", mrp.serviceName, mrp.cr.Namespace)
	selectorLabels := mrp.GetMonitoringSelectorLabels()
	monitoringCustomLabels := mrp.GetMonitoringCustomLabels(monitoringLabels)
	replicas := int32(1)
	volumes := mrp.getMonitoringVolumes()
	volumeMounts := mrp.getMonitoringVolumeMounts()
	ports := []corev1.ContainerPort{
		{
			ContainerPort: 8125,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			ContainerPort: 8094,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			ContainerPort: 8092,
			Protocol:      corev1.ProtocolUDP,
		},
	}
	envVars := mrp.getMonitoringEnvironmentVariables()
	envVars = append(envVars, mrp.getZooKeeperCredentialsEnvs()...)

	if mrp.spec.MonitoringType == "prometheus" {
		// Name is reduced because it must be no more than 15 characters
		ports = append(ports, corev1.ContainerPort{
			Name:          "prometheus-cli",
			ContainerPort: 8096,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	if IsVaultSecretManagementEnabled(mrp.cr) {
		envVars = append(envVars, getVaultConnectionEnvVars(mrp.GetServiceName(), mrp.cr)...)
		volumes = append(volumes, corev1.Volume{Name: "vault-env", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "vault-env", MountPath: "/vault"})
	}

	if mrp.cr.Spec.Global.ZooKeeperSsl.Enabled && mrp.cr.Spec.Global.ZooKeeperSsl.SecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "ssl-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: mrp.cr.Spec.Global.ZooKeeperSsl.SecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "ssl-certs", MountPath: "/tls"})
	}

	if mrp.cr.Spec.BackupDaemon != nil && mrp.cr.Spec.BackupDaemon.BackupDaemonSsl.Enabled {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BACKUP_DAEMON_TLS_ENABLED",
			Value: "true",
		})
		volumes = append(volumes, corev1.Volume{
			Name: "backup-daemon-tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: mrp.cr.Spec.BackupDaemon.BackupDaemonSsl.SecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "backup-daemon-tls-certs", MountPath: "/tls/backup"})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mrp.serviceName,
			Namespace: mrp.cr.Namespace,
			Labels:    monitoringLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: monitoringCustomLabels,
				},
				Spec: corev1.PodSpec{
					Volumes:           volumes,
					Affinity:          &mrp.spec.Affinity,
					Tolerations:       mrp.spec.Tolerations,
					PriorityClassName: mrp.spec.PriorityClassName,
					InitContainers:    mrp.getInitContainers(),
					Containers: []corev1.Container{
						{
							Name:            mrp.serviceName,
							Image:           mrp.spec.DockerImage,
							Ports:           ports,
							Env:             envVars,
							Resources:       mrp.spec.Resources,
							VolumeMounts:    volumeMounts,
							ImagePullPolicy: corev1.PullAlways,
							Command:         mrp.getCommand(),
							Args:            mrp.getArgs(),
							SecurityContext: getDefaultContainerSecurityContext(),
						},
					},
					SecurityContext:    &mrp.spec.SecurityContext,
					ServiceAccountName: mrp.GetServiceAccountName(),
					Hostname:           mrp.serviceName,
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
		},
	}
}

// getMonitoringVolumes configures the list of ZooKeeper Monitoring volumes
func (mrp MonitoringResourceProvider) getMonitoringVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-configuration", mrp.serviceName),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "config",
							Path: "telegraf.conf",
						},
					},
				},
			},
		},
	}
}

// getMonitoringVolumeMounts configures the list of ZooKeeper Monitoring volume mounts
func (mrp MonitoringResourceProvider) getMonitoringVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/telegraf",
		},
	}
}

// getMonitoringEnvironmentVariables configures the list of ZooKeeper Monitoring environment variables
func (mrp MonitoringResourceProvider) getMonitoringEnvironmentVariables() []corev1.EnvVar {
	environmentVariables := []corev1.EnvVar{
		{
			Name: "OS_PROJECT",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "ZOOKEEPER_HOST",
			Value: mrp.spec.ZooKeeperHost,
		},
		{
			Name:  "ZOOKEEPER_ENABLE_SSL",
			Value: strconv.FormatBool(mrp.cr.Spec.Global.ZooKeeperSsl.Enabled),
		},
	}
	if mrp.cr.Spec.BackupDaemon != nil {
		s3Enabled := mrp.cr.Spec.BackupDaemon.S3 != nil && mrp.cr.Spec.BackupDaemon.S3.Enabled
		environmentVariables = append(environmentVariables, []corev1.EnvVar{
			{
				Name:  "ZOOKEEPER_BACKUP_DAEMON_HOST",
				Value: mrp.spec.ZooKeeperBackupDaemonHost,
			},
			{
				Name:  "S3_ENABLED",
				Value: strconv.FormatBool(s3Enabled),
			},
			{
				Name:      "ZOOKEEPER_BACKUP_DAEMON_USERNAME",
				ValueFrom: getSecretEnvVarSource(mrp.cr.Spec.BackupDaemon.SecretName, "username"),
			},
			{
				Name:      "ZOOKEEPER_BACKUP_DAEMON_PASSWORD",
				ValueFrom: getSecretEnvVarSource(mrp.cr.Spec.BackupDaemon.SecretName, "password"),
			},
		}...)
	}
	return environmentVariables
}

func (mrp MonitoringResourceProvider) getCommand() []string {
	if IsVaultSecretManagementEnabled(mrp.cr) {
		return []string{"/vault/vault-env"}
	}
	return nil
}

func (mrp MonitoringResourceProvider) getArgs() []string {
	if IsVaultSecretManagementEnabled(mrp.cr) {
		return []string{"/docker-entrypoint.sh"}
	}
	return nil
}

func (mrp MonitoringResourceProvider) getInitContainers() []corev1.Container {
	if IsVaultSecretManagementEnabled(mrp.cr) {
		return []corev1.Container{
			getVaultInitContainer(mrp.cr),
		}
	}
	return nil
}

func (mrp MonitoringResourceProvider) getZooKeeperCredentialsEnvs() []corev1.EnvVar {
	if IsVaultSecretManagementEnabled(mrp.cr) {
		return []corev1.EnvVar{
			{
				Name:  "ZOOKEEPER_CLIENT_USERNAME",
				Value: getVaultSecretEnvVarSource(mrp.cr.Name, mrp.cr, "client-credentials", "username"),
			},
			{
				Name:  "ZOOKEEPER_CLIENT_PASSWORD",
				Value: getVaultSecretEnvVarSource(mrp.cr.Name, mrp.cr, "client-credentials", "password"),
			},
		}
	} else {
		return []corev1.EnvVar{
			{
				Name:      "ZOOKEEPER_CLIENT_USERNAME",
				ValueFrom: getSecretEnvVarSource(mrp.spec.SecretName, "zookeeper-client-username"),
			},
			{
				Name:      "ZOOKEEPER_CLIENT_PASSWORD",
				ValueFrom: getSecretEnvVarSource(mrp.spec.SecretName, "zookeeper-client-password"),
			},
		}
	}
}

// GetMonitoringLabels configures common labels for ZooKeeper Monitoring resources
func (mrp MonitoringResourceProvider) GetMonitoringLabels() map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = mrp.serviceName
	labels = util.JoinMaps(util.JoinMaps(labels, mrp.GetMonitoringSelectorLabels()), mrp.cr.Spec.Global.DefaultLabels)
	return labels
}

func (mrp MonitoringResourceProvider) GetMonitoringSelectorLabels() map[string]string {
	return map[string]string{
		"name":      mrp.serviceName,
		"component": "zookeeper-monitoring",
	}
}

func (mrp MonitoringResourceProvider) GetMonitoringCustomLabels(monitoringLabels map[string]string) map[string]string {
	globalLabels := mrp.cr.Spec.Global.CustomLabels
	customLabels := mrp.spec.CustomLabels
	return util.JoinMaps(util.JoinMaps(globalLabels, customLabels), monitoringLabels)
}

// GetServiceAccountName returns service account name for pods. Now it's equal to service name.
func (mrp MonitoringResourceProvider) GetServiceAccountName() string {
	return mrp.GetServiceName()
}
