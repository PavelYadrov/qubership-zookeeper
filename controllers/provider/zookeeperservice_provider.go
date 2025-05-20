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
	"github.com/Netcracker/qubership-zookeeper/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	SnapshotsPersistentVolumeClaimPattern = "pvc-%s-snapshots"
)

// GetZooKeeperLabels configures common labels for ZooKeeper resources
func GetZooKeeperLabels(serviceName string, defaultLabels map[string]string) map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = serviceName
	labels["name"] = serviceName
	labels = util.JoinMaps(util.JoinMaps(labels, GetZooKeeperSelectorLabels(serviceName)), defaultLabels)
	return labels
}

func GetZooKeeperSelectorLabels(serviceName string) map[string]string {
	return map[string]string{
		"component":   "zookeeper",
		"clusterName": serviceName,
	}
}

// getSecretEnvVarSource returns EnvVarSource for secret value
func getSecretEnvVarSource(secretName string, key string) *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			Key:                  key,
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		},
	}
}

// newServiceForCR returns service with specified parameters
func newServiceForCR(serviceName string, namespace string, labels map[string]string, selectorLabels map[string]string, ports []corev1.ServicePort) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports:       ports,
			Selector:    selectorLabels,
			ExternalIPs: nil,
		},
	}
}

// NewServiceAccount returns service account with specified parameters
func NewServiceAccount(serviceAccountName string, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: serviceAccountName, Namespace: namespace},
	}
}

// ProcessNonSharedPersistentVolumeClaim returns non-shared persistent volume claim according to the specified parameters.
func ProcessNonSharedPersistentVolumeClaim(persistentVolumeClaimName string, persistentVolumeName string,
	persistentVolumeLabel string, storageClassName *string, storageSize string, namespace string, labels map[string]string,
	logger logr.Logger) *corev1.PersistentVolumeClaim {
	var labelSelector *metav1.LabelSelector
	var detailsLog string
	if persistentVolumeName != "" {
		detailsLog = "volume name"
	} else if persistentVolumeLabel != "" {
		detailsLog = "label"
		keyValue := strings.Split(persistentVolumeLabel, "=")
		labelSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				keyValue[0]: keyValue[1],
			},
		}
	} else if storageClassName == nil {
		return nil
	}

	if storageClassName != nil && *storageClassName != "" {
		if detailsLog != "" {
			detailsLog = fmt.Sprintf("%s and ", detailsLog)
		}
		detailsLog = fmt.Sprintf("%sclass name", detailsLog)
	}

	logger.Info(fmt.Sprintf("Persistent volume claim [%s] is created by %s.", persistentVolumeClaimName, detailsLog))
	return NewPersistentVolumeClaim(persistentVolumeClaimName, namespace, labels, false, persistentVolumeName,
		labelSelector, storageClassName, storageSize)
}

// NewPersistentVolumeClaim configures persistent volume claim based on the specified parameters
func NewPersistentVolumeClaim(persistentVolumeClaimName string, namespace string, labels map[string]string, shared bool,
	persistentVolumeName string, labelSelector *metav1.LabelSelector, storageClassName *string, volumeSize string) *corev1.PersistentVolumeClaim {
	accessMode := corev1.ReadWriteOnce
	if shared {
		accessMode = corev1.ReadWriteMany
	}

	persistentVolumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      persistentVolumeClaimName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(volumeSize),
				},
			},
		},
	}

	if persistentVolumeName != "" {
		persistentVolumeClaim.Spec.VolumeName = persistentVolumeName
	} else if labelSelector != nil {
		persistentVolumeClaim.Spec.Selector = labelSelector
	} else {
		persistentVolumeClaim.Spec.StorageClassName = storageClassName
	}

	if persistentVolumeName != "" || labelSelector != nil {
		if storageClassName != nil {
			persistentVolumeClaim.Spec.StorageClassName = storageClassName
		} else {
			persistentVolumeClaim.Spec.StorageClassName = new(string)
		}
	}

	return persistentVolumeClaim
}

// buildEnvs builds array of specified environment variables with additional list of environment variables
func buildEnvs(envVars []corev1.EnvVar, additionalEnvs []string, logger logr.Logger) []corev1.EnvVar {
	for _, envVar := range additionalEnvs {
		envPair := strings.SplitN(envVar, "=", 2)
		if len(envPair) == 2 {
			if name := strings.TrimSpace(envPair[0]); len(name) > 0 {
				value := strings.TrimSpace(envPair[1])
				envVars = append(envVars, corev1.EnvVar{Name: name, Value: value})
				continue
			}
		}
		logger.Info(fmt.Sprintf("Environment variable \"%s\" is incorrect", envVar))
	}
	return envVars
}

// getDefaultContainerSecurityContext returns default security context for containers for deployment to restricted environment
func getDefaultContainerSecurityContext() *corev1.SecurityContext {
	falseValue := false
	return &corev1.SecurityContext{AllowPrivilegeEscalation: &falseValue,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	}
}
