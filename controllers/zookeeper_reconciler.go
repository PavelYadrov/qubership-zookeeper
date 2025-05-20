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
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	zooKeeperConditionReason = "ZooKeeperReadinessStatus"
	zooKeeperHashName        = "spec.zookeeper"
)

type ReconcileZooKeeper struct {
	reconciler *ZooKeeperServiceReconciler
	cr         *zookeeperservice.ZooKeeperService
	logger     logr.Logger
	zkProvider provider.ZooKeeperResourceProvider
}

func (r ReconcileZooKeeper) Status() error {
	if err := r.reconciler.updateConditions(r.cr,
		NewCondition(statusFalse,
			typeInProgress,
			zooKeeperConditionReason,
			"ZooKeeper health check")); err != nil {
		return err
	}
	r.logger.Info("Start checking for ZooKeeper pods")
	err := wait.PollImmediate(10*time.Second, time.Duration(r.cr.Spec.Global.PodsReadyTimeout)*time.Second, func() (done bool, err error) {
		for i := 1; i <= r.cr.Spec.ZooKeeper.Replicas; i++ {
			deploymentName := fmt.Sprintf("%s-%d", r.cr.Name, i)
			if !r.reconciler.isDeploymentReady(deploymentName, r.cr.Namespace, r.logger) {
				r.logger.Info(fmt.Sprintf("%s is not ready yet", deploymentName))
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return r.reconciler.updateConditions(r.cr, NewCondition(statusFalse,
			typeFailed,
			zooKeeperConditionReason,
			"ZooKeeper pods are not ready"))
	}
	return r.reconciler.updateConditions(r.cr, NewCondition(statusTrue,
		typeReady,
		zooKeeperConditionReason,
		"ZooKeeper pods are ready"))
}

func NewReconcileZooKeeper(r *ZooKeeperServiceReconciler, cr *zookeeperservice.ZooKeeperService, logger logr.Logger) ReconcileZooKeeper {
	return ReconcileZooKeeper{
		reconciler: r,
		cr:         cr,
		logger:     logger,
		zkProvider: provider.NewZooKeeperResourceProvider(cr, logger),
	}
}

func (r ReconcileZooKeeper) Reconcile() error {
	// Find the secret and add a reference to it
	zooKeeperSecret, err := r.reconciler.watchSecret(r.cr.Spec.ZooKeeper.SecretName, r.cr, r.logger)
	if err != nil {
		if provider.IsVaultSecretManagementEnabled(r.cr) && errors.IsNotFound(err) {
			log.Info("Secret has been removed. Cannot watch for it to rotate secrets")
			zooKeeperSecret = &corev1.Secret{}
		} else {
			return err
		}
	}

	zooKeeperSpecHash, err := util.Hash(r.cr.Spec.ZooKeeper)
	if err != nil {
		return err
	}
	if r.reconciler.ResourceHashes[zooKeeperHashName] == zooKeeperSpecHash &&
		r.reconciler.ResourceHashes[globalHashName] == globalSpecHash &&
		(zooKeeperSecret.Name == "" || r.reconciler.ResourceVersions[zooKeeperSecret.Name] == zooKeeperSecret.ResourceVersion) {
		r.logger.Info("ZooKeeper configuration didn't change, skipping reconcile loop")
		return nil
	}
	zkProvider := r.zkProvider
	zookeeperSpec := r.cr.Spec.ZooKeeper
	if zookeeperSpec.Replicas > 0 {
		// Create snapshots persistent volume claim if SnapshotStorage.PersistentVolumeType is not empty
		if zookeeperSpec.SnapshotStorage.PersistentVolumeType != "" && zookeeperSpec.SnapshotStorage.PersistentVolumeType != "standalone" {
			snapshotPersistentVolumeClaim, err := r.reconciler.processSnapshotsPersistentVolumeClaim(zookeeperSpec.SnapshotStorage, r.cr, r.logger)
			if err != nil {
				return err
			}
			if snapshotPersistentVolumeClaim != nil {
				if err := r.reconciler.createPersistentVolumeClaim(snapshotPersistentVolumeClaim, r.logger); err != nil {
					return err
				}
			}
		}

		// Define a new client service object
		clientService := zkProvider.NewZooKeeperClientServiceForCR()
		if err := controllerutil.SetControllerReference(r.cr, clientService, r.reconciler.Scheme); err != nil {
			return err
		}
		if err := r.reconciler.createOrUpdateService(clientService, r.logger); err != nil {
			return err
		}
		// Define a new domain service object
		domainService := zkProvider.NewZooKeeperDomainServiceForCR()
		if err := controllerutil.SetControllerReference(r.cr, domainService, r.reconciler.Scheme); err != nil {
			return err
		}
		if err := r.reconciler.createOrUpdateService(domainService, r.logger); err != nil {
			return err
		}

		currentReplicas, err := r.getCurrentDeploymentsCount()
		if err != nil {
			return err
		}

		if currentReplicas <= 2 || currentReplicas != zookeeperSpec.Replicas {
			r.logger.Info("RollingUpdate value set to false")
			r.cr.Spec.ZooKeeper.RollingUpdate = false
		}

		if currentReplicas > zookeeperSpec.Replicas {
			r.logger.Info(fmt.Sprintf("There is an attempt to downscale ZooKeeper with %d replicas to ZooKeeper with %d replicas. For correct work excess ZooKeeper deployments need to be scaled down.", currentReplicas, zookeeperSpec.Replicas))
			for i := zookeeperSpec.Replicas + 1; i <= currentReplicas; i++ {
				if err := r.reconciler.scaleDeployment(fmt.Sprintf("%s-%d", r.cr.Name, i), 0, r.cr.Namespace, r.logger); err != nil {
					return err
				}
			}
		}

		for serverId := 1; serverId <= zookeeperSpec.Replicas; serverId++ {
			// Define a new server Service object
			serverService := zkProvider.NewZooKeeperServerServiceForCR(serverId)
			if err := controllerutil.SetControllerReference(r.cr, serverService, r.reconciler.Scheme); err != nil {
				return err
			}
			if err := r.reconciler.createOrUpdateService(serverService, r.logger); err != nil {
				return err
			}

			// Define a new PersistentVolumeClaim object
			persistentVolumeClaim := zkProvider.NewZooKeeperPersistentVolumeClaimForCR(serverId)
			if persistentVolumeClaim != nil {
				if err := r.reconciler.createPersistentVolumeClaim(persistentVolumeClaim, r.logger); err != nil {
					return err
				}
			}

			serviceAccount := provider.NewServiceAccount(zkProvider.GetServiceAccountName(), r.cr.Namespace)
			if err := r.reconciler.createServiceAccount(serviceAccount, r.logger); err != nil {
				return err
			}

			if provider.IsVaultSecretManagementEnabled(r.cr) {
				err := r.processVaultSecrets(zooKeeperSecret)
				if err != nil {
					return err
				}
			}

			// Define a new Deployment object
			serverDeployment := zkProvider.NewServerDeploymentForCR(serverId)
			if err := controllerutil.SetControllerReference(r.cr, serverDeployment, r.reconciler.Scheme); err != nil {
				return err
			}
			if err := r.reconciler.createOrUpdateDeployment(serverDeployment, r.logger); err != nil {
				return err
			}

			//Checking for pod to be in running state
			deploymentName := serverDeployment.Name
			r.logger.Info(fmt.Sprintf("Waiting for pod of %s deployment to be in 'Running' state.", deploymentName))
			err = wait.Poll(waitingInterval, time.Duration(300)*time.Second, func() (done bool, err error) {
				podRunning, err := r.isPodRunning(r.cr, deploymentName)
				if err != nil {
					r.logger.Error(err, "Error checking if pod is running.")
					return false, err
				}
				r.logger.Info("Pod is ready!")

				return podRunning, nil
			})

			if err != nil {
				r.logger.Error(err, fmt.Sprintf("Pod for deployment %s is not in 'Running' state within the expected time.", deploymentName))
				return err
			}

			if r.cr.Spec.ZooKeeper.RollingUpdate {
				deploymentName := fmt.Sprintf("%s-%d", r.cr.Name, serverId)
				r.logger.Info(fmt.Sprintf("Waiting for %s deployment.", deploymentName))
				time.Sleep(waitingInterval)
				err = wait.PollImmediate(waitingInterval, time.Duration(300)*time.Second, func() (done bool, err error) {
					return r.reconciler.isDeploymentReady(deploymentName, r.cr.Namespace, r.logger), nil
				})
				if err != nil {
					r.logger.Error(err, fmt.Sprintf("Deployment %s failed.", deploymentName))
					return err
				}
			}

		}
	}
	r.logger.Info("Updating ZooKeeper status")
	if err := r.updateZooKeeperStatus(r.cr); err != nil {
		return err
	}

	r.reconciler.ResourceHashes[zooKeeperHashName] = zooKeeperSpecHash
	r.reconciler.ResourceVersions[zooKeeperSecret.Name] = zooKeeperSecret.ResourceVersion
	return nil
}

func (r *ReconcileZooKeeper) getCurrentDeploymentsCount() (int, error) {
	deployments, err := r.findZookeperDeployments(r.cr)
	if err != nil {
		return 0, err
	}
	var activeDeploymentsCount int
	for _, deployment := range deployments.Items {
		if *deployment.Spec.Replicas > 0 {
			activeDeploymentsCount++
		}
	}
	return activeDeploymentsCount, nil
}

func (r *ReconcileZooKeeper) isPodRunning(cr *zookeeperservice.ZooKeeperService, deploymentName string) (bool, error) {
	pods, err := r.getPodsForDeployment(deploymentName, cr)
	if err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			r.logger.Info("Pod is not in running state.")
			return false, nil
		}
	}
	return true, nil
}

func (r *ReconcileZooKeeper) findZookeperDeployments(cr *zookeeperservice.ZooKeeperService) (*appsv1.DeploymentList, error) {
	zookeperLabels := provider.GetZooKeeperSelectorLabels(cr.Name)
	return r.reconciler.findDeploymentList(cr.Namespace, zookeperLabels)
}

// updateZooKeeperStatus updates the ZooKeeper status
func (r *ReconcileZooKeeper) updateZooKeeperStatus(cr *zookeeperservice.ZooKeeperService) error {
	labels := provider.GetZooKeeperSelectorLabels(r.cr.Name)
	foundPodList, err := r.reconciler.findPodList(r.cr.Namespace, labels)
	if err != nil {
		return err
	}
	r.cr.Status.ZooKeeperStatus.Servers = getPodNames(foundPodList.Items)
	return r.reconciler.Client.Status().Update(context.TODO(), cr)
}
