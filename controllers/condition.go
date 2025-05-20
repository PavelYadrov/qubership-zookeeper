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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	statusTrue      = "True"
	statusFalse     = "False"
	typeSuccessful  = "Successful"
	typeReady       = "Ready"
	typeFailed      = "Failed"
	typeInProgress  = "In progress"
	waitingInterval = 10 * time.Second
)

func NewCondition(conditionStatus string, conditionType string, conditionReason string, conditionMessage string) zookeeperservice.StatusCondition {
	return zookeeperservice.StatusCondition{
		Type:    conditionType,
		Status:  conditionStatus,
		Reason:  conditionReason,
		Message: conditionMessage,
	}
}

func (r *ZooKeeperServiceReconciler) updateConditions(cr *zookeeperservice.ZooKeeperService, condition zookeeperservice.StatusCondition) error {
	currentConditions := cr.Status.Conditions
	condition.LastTransitionTime = metav1.Now().String()
	currentConditions = addCondition(currentConditions, condition)

	cr.Status.Conditions = currentConditions
	log.Info(fmt.Sprintf("Update condition status: %+v", condition))
	return r.Client.Status().Update(context.TODO(), cr)
}

func addCondition(currentConditions []zookeeperservice.StatusCondition, condition zookeeperservice.StatusCondition) []zookeeperservice.StatusCondition {
	for i, currentCondition := range currentConditions {
		if condition.Reason == currentCondition.Reason {
			if condition.Type != currentCondition.Type ||
				condition.Status != currentCondition.Status ||
				condition.Message != currentCondition.Message {
				currentConditions[i] = condition
			}
			return currentConditions
		}
	}
	return append(currentConditions, condition)
}

func hasFailedConditions(cr *zookeeperservice.ZooKeeperService) bool {
	for _, condition := range cr.Status.Conditions {
		if condition.Type == "Failed" {
			return true
		}
	}
	return false
}
