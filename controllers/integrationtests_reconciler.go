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
	zookeeperservice "github.com/Netcracker/qubership-zookeeper/api/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

const integrationTestsConditionReason = "ZooKeeperIntegrationTestsStatus"

type ReconcileIntegrationTests struct {
	reconciler *ZooKeeperServiceReconciler
	cr         *zookeeperservice.ZooKeeperService
	logger     logr.Logger
}

func NewReconcileIntegrationTests(r *ZooKeeperServiceReconciler, cr *zookeeperservice.ZooKeeperService, logger logr.Logger) ReconcileIntegrationTests {
	return ReconcileIntegrationTests{
		reconciler: r,
		cr:         cr,
		logger:     logger,
	}
}

func (r ReconcileIntegrationTests) Status() error {
	if !r.cr.Spec.IntegrationTests.WaitForResult {
		return nil
	}

	if err := r.reconciler.updateConditions(r.cr,
		NewCondition(statusFalse,
			typeInProgress,
			integrationTestsConditionReason,
			"Start checking for ZooKeeper Integration Tests")); err != nil {
		return err
	}
	r.logger.Info("Start checking for ZooKeeper Integration Tests")
	err := wait.PollImmediate(10*time.Second, time.Duration(r.cr.Spec.IntegrationTests.Timeout)*time.Second, func() (done bool, err error) {
		if r.reconciler.isDeploymentReady(r.cr.Spec.IntegrationTests.ServiceName, r.cr.Namespace, r.logger) {
			return true, nil
		}
		r.logger.Info("ZooKeeper Integration Tests deployment is not ready yet")
		return false, nil
	})
	if err != nil {
		return r.reconciler.updateConditions(r.cr, NewCondition(statusFalse,
			typeFailed,
			integrationTestsConditionReason,
			"ZooKeeper Integration Tests failed. See more details in integration test logs."))
	}
	return r.reconciler.updateConditions(r.cr, NewCondition(statusTrue,
		typeReady,
		integrationTestsConditionReason,
		"ZooKeeper Integration Tests performed successfully"))
}

func (r ReconcileIntegrationTests) Reconcile() error {
	return nil
}
