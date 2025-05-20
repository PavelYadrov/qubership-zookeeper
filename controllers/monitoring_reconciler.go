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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	monitoringConditionReason = "ZooKeeperMonitoringReadinessStatus"
	monitoringHashName        = "spec.monitoring"
)

type ReconcileMonitoring struct {
	reconciler         *ZooKeeperServiceReconciler
	cr                 *zookeeperservice.ZooKeeperService
	logger             logr.Logger
	monitoringProvider provider.MonitoringResourceProvider
}

func (r ReconcileMonitoring) Status() error {
	if err := r.reconciler.updateConditions(r.cr,
		NewCondition(statusFalse,
			typeInProgress,
			monitoringConditionReason,
			"ZooKeeper Monitoring health check")); err != nil {
		return err
	}
	r.logger.Info("Start checking for ZooKeeper monitoring pod")
	err := wait.PollImmediate(10*time.Second, time.Duration(r.cr.Spec.Global.PodsReadyTimeout)*time.Second, func() (done bool, err error) {
		if r.reconciler.isDeploymentReady(r.monitoringProvider.GetServiceName(), r.cr.Namespace, r.logger) {
			return true, nil
		}
		r.logger.Info(fmt.Sprintf("%s is not ready yet", r.monitoringProvider.GetServiceName()))
		return false, nil
	})
	if err != nil {
		return r.reconciler.updateConditions(r.cr, NewCondition(statusFalse,
			typeFailed,
			monitoringConditionReason,
			"ZooKeeper Monitoring pod is not ready"))
	}
	return r.reconciler.updateConditions(r.cr, NewCondition(statusTrue,
		typeReady,
		monitoringConditionReason,
		"ZooKeeper Monitoring pod is ready"))
}

func NewReconcileMonitoring(r *ZooKeeperServiceReconciler, cr *zookeeperservice.ZooKeeperService, logger logr.Logger) ReconcileMonitoring {
	return ReconcileMonitoring{
		reconciler:         r,
		cr:                 cr,
		logger:             logger,
		monitoringProvider: provider.NewMonitoringResourceProvider(cr, logger),
	}
}

func (r ReconcileMonitoring) Reconcile() error {

	monitoringSecret, err := r.reconciler.watchSecret(r.cr.Spec.Monitoring.SecretName, r.cr, r.logger)
	if err != nil {
		if provider.IsVaultSecretManagementEnabled(r.cr) && errors.IsNotFound(err) {
			log.Info("Secret has been removed. Cannot watch for it to rotate secrets")
			monitoringSecret = &corev1.Secret{}
		} else {
			return err
		}
	}

	monitoringSpecHash, err := util.Hash(r.cr.Spec.Monitoring)
	if err != nil {
		return err
	}
	if r.reconciler.ResourceHashes[monitoringHashName] == monitoringSpecHash &&
		r.reconciler.ResourceHashes[globalHashName] == globalSpecHash &&
		(monitoringSecret.Name == "" || r.reconciler.ResourceVersions[monitoringSecret.Name] == monitoringSecret.ResourceVersion) {
		r.logger.Info("ZooKeeper Monitoring configuration didn't change, skipping reconcile loop")
		return nil
	}

	clientService := r.monitoringProvider.NewMonitoringClientService()
	if err := controllerutil.SetControllerReference(r.cr, clientService, r.reconciler.Scheme); err != nil {
		return err
	}
	if err := r.reconciler.createOrUpdateService(clientService, r.logger); err != nil {
		return err
	}

	serviceAccount := provider.NewServiceAccount(r.monitoringProvider.GetServiceAccountName(), r.cr.Namespace)
	if err := r.reconciler.createServiceAccount(serviceAccount, r.logger); err != nil {
		return err
	}

	if provider.IsVaultSecretManagementEnabled(r.cr) {
		err := r.processVaultSecrets(monitoringSecret)
		if err != nil {
			return err
		}
	}

	deployment := r.monitoringProvider.NewMonitoringDeployment()
	if err := controllerutil.SetControllerReference(r.cr, deployment, r.reconciler.Scheme); err != nil {
		return err
	}
	if err := r.reconciler.createOrUpdateDeployment(deployment, r.logger); err != nil {
		return err
	}

	r.logger.Info("Updating ZooKeeper Monitoring status")
	if err := r.updateMonitoringStatus(r.cr); err != nil {
		return err
	}

	r.reconciler.ResourceHashes[monitoringHashName] = monitoringSpecHash
	r.reconciler.ResourceVersions[monitoringSecret.Name] = monitoringSecret.ResourceVersion
	return nil
}

// updateMonitoringStatus updates the status of ZooKeeper Monitoring
func (r *ReconcileMonitoring) updateMonitoringStatus(cr *zookeeperservice.ZooKeeperService) error {
	labels := r.monitoringProvider.GetMonitoringSelectorLabels()
	foundPodList, err := r.reconciler.findPodList(r.cr.Namespace, labels)
	if err != nil {
		return err
	}
	r.cr.Status.MonitoringStatus.Nodes = getPodNames(foundPodList.Items)
	return r.reconciler.Client.Status().Update(context.TODO(), cr)
}
