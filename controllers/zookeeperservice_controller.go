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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	zooKeeperServiceConditionReason = "ReconcileCycleStatus"
	globalHashName                  = "spec.global"
)

var (
	log            = logf.Log.WithName("controller_zookeeperservice")
	globalSpecHash = ""
)

type ReconcileService interface {
	Reconcile() error
	Status() error
}

//+kubebuilder:rbac:groups=qubership.org,resources=zookeeperservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubership.org,resources=zookeeperservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubership.org,resources=zookeeperservices/finalizers,verbs=update

func (r *ZooKeeperServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ZooKeeper Service")

	// Fetch the ZooKeeperService instance
	instance := &zookeeperservice.ZooKeeperService{}
	if err := r.Client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	specHash, err := util.Hash(instance.Spec)
	if err != nil {
		reqLogger.Info("error in hash function")
		return reconcile.Result{}, err
	}
	globalSpecHash, err = util.Hash(instance.Spec.Global)
	if err != nil {
		reqLogger.Info("error in hash function for global section")
		return reconcile.Result{}, err
	}
	isCustomResourceChanged := r.ResourceHashes["spec"] != specHash
	if isCustomResourceChanged {
		instance.Status.Conditions = []zookeeperservice.StatusCondition{}
		if err := r.updateConditions(instance, NewCondition(statusFalse,
			typeInProgress,
			zooKeeperServiceConditionReason,
			"Reconciliation cycle started")); err != nil {
			return reconcile.Result{}, err
		}
	}

	if provider.IsVaultSecretManagementEnabled(instance) {
		if err := r.InitVaultClient(instance); err != nil {
			r.writeFailedStatus(instance, fmt.Sprintf("An error occurred while creating Vault client: %v", err))
			return reconcile.Result{}, err
		}
	}

	reconcilers := r.buildReconcilers(instance, log)

	for _, reconciler := range reconcilers {
		if err := reconciler.Reconcile(); err != nil {
			reqLogger.Error(err, fmt.Sprintf("Error when reconciling `%v`", reconciler))
			r.writeFailedStatus(instance, fmt.Sprintf("Reconciliation cycle failed for %T due to: %v", reconciler, err))
			return reconcile.Result{}, err
		}
	}

	if isCustomResourceChanged {
		if instance.Spec.Global != nil && instance.Spec.Global.WaitForPodsReady {
			if err := r.updateConditions(instance, NewCondition(statusFalse,
				typeInProgress,
				zooKeeperServiceConditionReason,
				"Checking deployment readiness status")); err != nil {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Waiting for resources to be applied")
			time.Sleep(10 * time.Second)

			for _, reconciler := range reconcilers {
				if err := reconciler.Status(); err != nil {
					r.writeFailedStatus(instance, fmt.Sprintf("The status reconciliation cycle failed for %T due to: %v", reconciler, err))
					return reconcile.Result{}, err
				}
			}
		}

		if hasFailedConditions(instance) {
			if err := r.updateConditions(instance, NewCondition(statusFalse,
				typeFailed,
				zooKeeperServiceConditionReason,
				"The deployment readiness status check failed")); err != nil {
				return reconcile.Result{}, err
			}
		} else {
			if err := r.updateConditions(instance,
				NewCondition(statusTrue,
					typeSuccessful,
					zooKeeperServiceConditionReason,
					"The deployment readiness status check is successful")); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	reqLogger.Info("Reconciliation cycle succeeded")
	r.ResourceHashes["spec"] = specHash
	r.ResourceHashes[globalHashName] = globalSpecHash
	return reconcile.Result{}, nil
}

func (r *ZooKeeperServiceReconciler) writeFailedStatus(instance *zookeeperservice.ZooKeeperService, errorMessage string) {
	if err := r.updateConditions(instance,
		NewCondition(statusFalse,
			typeFailed,
			zooKeeperServiceConditionReason,
			errorMessage)); err != nil {
		log.Error(err, "An error occurred while updating the status condition")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZooKeeperServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	statusPredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&zookeeperservice.ZooKeeperService{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(statusPredicate).
		Complete(r)
}

// buildReconcilers returns service reconcilers in accordance with custom resource.
func (r *ZooKeeperServiceReconciler) buildReconcilers(cr *zookeeperservice.ZooKeeperService, logger logr.Logger) []ReconcileService {
	var reconcilers []ReconcileService
	if cr.Spec.ZooKeeper != nil {
		reconcilers = append(reconcilers, NewReconcileZooKeeper(r, cr, logger))
	}
	if cr.Spec.Monitoring != nil {
		reconcilers = append(reconcilers, NewReconcileMonitoring(r, cr, logger))
	}
	if cr.Spec.BackupDaemon != nil {
		reconcilers = append(reconcilers, NewReconcileBackupDaemon(r, cr, logger))
	}
	if cr.Spec.IntegrationTests != nil {
		reconcilers = append(reconcilers, NewReconcileIntegrationTests(r, cr, logger))
	}
	return reconcilers
}
