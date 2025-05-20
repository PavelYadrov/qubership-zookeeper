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

package controllers

import (
	"context"
	"fmt"
	zookeeperservice "github.com/Netcracker/qubership-zookeeper/api/v1"
	"github.com/Netcracker/qubership-zookeeper/controllers/provider"
	"github.com/Netcracker/qubership-zookeeper/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ZooKeeperServiceReconciler reconciles a ZooKeeperService object
type ZooKeeperServiceReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client           client.Client
	Scheme           *runtime.Scheme
	ResourceVersions map[string]string
	ResourceHashes   map[string]string
}

// createOrUpdateService creates the service if it doesn't exist and updates otherwise
func (r *ZooKeeperServiceReconciler) createOrUpdateService(service *corev1.Service, logger logr.Logger) error {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] service", service.Name))
	foundService := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new service",
			"Service.Namespace", service.Namespace, "Service.Name", service.Name)
		return r.Client.Create(context.TODO(), service)
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating the found service",
			"Service.Namespace", service.Namespace, "Service.Name", service.Name)
		service.ResourceVersion = foundService.ResourceVersion
		if foundService.Spec.Type == corev1.ServiceTypeClusterIP {
			service.Spec.ClusterIP = foundService.Spec.ClusterIP
		}
		return r.Client.Update(context.TODO(), service)
	}
}

// createServiceAccount
func (r *ZooKeeperServiceReconciler) createServiceAccount(serviceAccount *corev1.ServiceAccount, logger logr.Logger) error {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] service account", serviceAccount.Name))
	foundService := &corev1.ServiceAccount{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name: serviceAccount.Name, Namespace: serviceAccount.Namespace,
	}, foundService)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new service account",
			"ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = r.Client.Create(context.TODO(), serviceAccount)
	}
	return err
}

// createPersistentVolumeClaim creates persistent volume claim if it does not exist; returns an error if any operation failed
func (r *ZooKeeperServiceReconciler) createPersistentVolumeClaim(persistentVolumeClaim *corev1.PersistentVolumeClaim, logger logr.Logger) error {
	// There is no ability to update PVC
	logger.Info(fmt.Sprintf("Checking Existence of [%s] persistent volume claim", persistentVolumeClaim.Name))
	foundPersistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(context.TODO(),
		types.NamespacedName{Name: persistentVolumeClaim.Name, Namespace: persistentVolumeClaim.Namespace},
		foundPersistentVolumeClaim)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new persistent volume claim",
			"PersistentVolumeClaim.Namespace", persistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", persistentVolumeClaim.Name)
		err = r.Client.Create(context.TODO(), persistentVolumeClaim)
	}
	return err
}

func (r *ZooKeeperServiceReconciler) findPersistentVolumeClaim(name string, namespace string, logger logr.Logger) (*corev1.PersistentVolumeClaim, error) {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] persistent volume claim", name))
	foundPersistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace},
		foundPersistentVolumeClaim)
	return foundPersistentVolumeClaim, err
}

// createOrUpdateDeployment creates deployment if it does not exist, or updates if it exists;
// returns an error if any operation failed
func (r *ZooKeeperServiceReconciler) createOrUpdateDeployment(deployment *appsv1.Deployment, logger logr.Logger) error {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] deployment", deployment.Name))
	foundDeployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(),
		types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace},
		foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new deployment",
			"Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return r.Client.Create(context.TODO(), deployment)
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating the found deployment",
			"Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return r.Client.Update(context.TODO(), deployment)
	}
}

func (r *ZooKeeperServiceReconciler) findDeployment(name string, namespace string, logger logr.Logger) (*appsv1.Deployment, error) {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] deployment", name))
	foundDeployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace},
		foundDeployment)
	return foundDeployment, err
}

func (r *ZooKeeperServiceReconciler) findDeploymentList(namespace string, deploymentLabels map[string]string) (*appsv1.DeploymentList, error) {
	foundDeploymentList := &appsv1.DeploymentList{}
	err := r.Client.List(context.TODO(), foundDeploymentList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(deploymentLabels),
	})
	return foundDeploymentList, err
}

// updateSecret updates secret if it exists; returns an error if any operation failed
func (r *ZooKeeperServiceReconciler) updateSecret(secret *corev1.Secret, logger logr.Logger) error {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] secret", secret.Name))
	foundSecret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(),
		types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace},
		foundSecret)
	if err != nil && errors.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("Secret [%s] must exist", secret.Name))
		return err
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating the found secret",
			"Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		return r.Client.Update(context.TODO(), secret)
	}
}

// findSecret finds secret by name
func (r *ZooKeeperServiceReconciler) findSecret(name string, namespace string, logger logr.Logger) (*corev1.Secret, error) {
	logger.Info(fmt.Sprintf("Checking Existence of [%s] secret", name))
	foundSecret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace},
		foundSecret)
	return foundSecret, err
}

// cleanSecretData
func (r *ZooKeeperServiceReconciler) cleanSecretData(secret *corev1.Secret, logger logr.Logger) error {
	logger.Info(fmt.Sprintf("Cleaning data of [%s] secret", secret.Name))
	secret.Data = nil
	return r.Client.Update(context.TODO(), secret)
}

// findPodList finds pods by labels
func (r *ZooKeeperServiceReconciler) findPodList(namespace string, podLabels map[string]string) (*corev1.PodList, error) {
	foundPodList := &corev1.PodList{}
	err := r.Client.List(context.TODO(), foundPodList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(podLabels),
	})
	return foundPodList, err
}

func (r *ReconcileZooKeeper) getPodsForDeployment(deploymentName string, cr *zookeeperservice.ZooKeeperService) (*corev1.PodList, error) {
	zookeperLabels := provider.GetZooKeeperLabels(cr.Name, cr.Spec.Global.DefaultLabels)
	pods, err := r.reconciler.findPodList(cr.Namespace, zookeperLabels)
	if err != nil {
		return nil, err
	}

	listOpts := []client.ListOption{
		client.MatchingLabels(map[string]string{"name": deploymentName}),
	}
	if err := r.reconciler.Client.List(context.TODO(), pods, listOpts...); err != nil {
		return nil, err
	}
	return pods, nil
}

// watchSecret finds secret by name and set current custom resource as owner to watch changes
func (r *ZooKeeperServiceReconciler) watchSecret(secretName string, cr *zookeeperservice.ZooKeeperService, logger logr.Logger) (*corev1.Secret, error) {
	secret, err := r.findSecret(secretName, cr.Namespace, logger)
	if err != nil {
		return nil, err
	} else {
		// Check if there's an existing owner reference
		if existing := metav1.GetControllerOf(secret); existing != nil && !referSameObject(existing.Name, existing.APIVersion, existing.Kind, cr.Name, cr.APIVersion, cr.Kind) {
			secret.OwnerReferences = nil
		}
		if err := controllerutil.SetControllerReference(cr, secret, r.Scheme); err != nil {
			return nil, err
		}
		if err := r.updateSecret(secret, logger); err != nil {
			return nil, err
		}
	}
	return secret, nil
}

func referSameObject(aName string, aGroup string, aKind string, bName string, bGroup string, bKind string) bool {
	aGV, err := schema.ParseGroupVersion(aGroup)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(bGroup)
	if err != nil {
		return false
	}

	return aGV.Group == bGV.Group && aKind == bKind && aName == bName
}

// processSnapshotsPersistentVolumeClaim returns snapshots persistent volume claim according to the PersistentVolumeType.
// Creates PersistenceVolume if necessary.
func (r *ZooKeeperServiceReconciler) processSnapshotsPersistentVolumeClaim(snapshotStorage zookeeperservice.SnapshotStorage,
	cr *zookeeperservice.ZooKeeperService, logger logr.Logger) (*corev1.PersistentVolumeClaim, error) {
	persistentVolumeClaimName := snapshotStorage.PersistentVolumeClaimName
	if persistentVolumeClaimName == "" {
		persistentVolumeClaimName = fmt.Sprintf(provider.SnapshotsPersistentVolumeClaimPattern, cr.Name)
		logger.Info(fmt.Sprintf("Parameter 'persistentVolumeClaimName' for snapshots storage is not specified, so default value '%s' is used",
			persistentVolumeClaimName))
	}
	snapshotsLabels := provider.GetZooKeeperLabels(cr.Name, cr.Spec.Global.DefaultLabels)

	if snapshotStorage.PersistentVolumeType == "standalone" {
		return provider.ProcessNonSharedPersistentVolumeClaim(persistentVolumeClaimName, snapshotStorage.PersistentVolumeName,
			snapshotStorage.PersistentVolumeLabel, snapshotStorage.StorageClass, snapshotStorage.VolumeSize,
			cr.Namespace, snapshotsLabels, logger), nil
	} else if snapshotStorage.PersistentVolumeType == "predefined_claim" {
		return r.findPersistentVolumeClaim(persistentVolumeClaimName, cr.Namespace, logger)
	} else if snapshotStorage.PersistentVolumeType == "predefined" {
		if snapshotStorage.PersistentVolumeName == "" {
			return nil, fmt.Errorf("parameter 'persistentVolumeName' must be specified for 'predefined' persistent volume type")
		}
		return provider.NewPersistentVolumeClaim(persistentVolumeClaimName, cr.Namespace, snapshotsLabels, true,
			snapshotStorage.PersistentVolumeName, nil, snapshotStorage.StorageClass, snapshotStorage.VolumeSize), nil
	} else if snapshotStorage.PersistentVolumeType == "storage_class" {
		return provider.NewPersistentVolumeClaim(persistentVolumeClaimName, cr.Namespace, snapshotsLabels, true,
			"", nil, snapshotStorage.StorageClass, snapshotStorage.VolumeSize), nil
	}
	return nil, nil
}

func (r *ZooKeeperServiceReconciler) isDeploymentReady(deploymentName string, namespace string, logger logr.Logger) bool {
	deployment, err := r.findDeployment(deploymentName, namespace, logger)
	if err != nil {
		logger.Error(err, "Cannot check deployment status")
		return false
	}
	availableReplicas := util.Min(deployment.Status.ReadyReplicas, deployment.Status.UpdatedReplicas)
	return *deployment.Spec.Replicas == availableReplicas
}

func secretContainsKey(zooKeeperSecret *corev1.Secret, key string) bool {
	return zooKeeperSecret.Data != nil &&
		zooKeeperSecret.Data[key] != nil &&
		string(zooKeeperSecret.Data[key]) != ""
}

// getPodNames returns the array of pod names
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Labels["name"])
	}
	return podNames
}

func (r *ZooKeeperServiceReconciler) scaleDeployment(name string, replicas int32, namespace string, logger logr.Logger) error {
	logger.Info(fmt.Sprintf("Scaling [%s] deployment to [%d] replicas", name, replicas))
	foundDeployment, err := r.findDeployment(name, namespace, logger)
	if err == nil {
		foundDeployment.Spec.Replicas = &replicas
		return r.Client.Update(context.TODO(), foundDeployment)
	}
	return err
}
