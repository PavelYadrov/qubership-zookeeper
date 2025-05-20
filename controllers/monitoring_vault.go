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
	"fmt"
	"github.com/Netcracker/qubership-zookeeper/controllers/provider"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcileMonitoring) processVaultSecrets(monitoringSecret *corev1.Secret) error {
	log.Info("Process vault secrets management for ZooKeeper Monitoring")
	if r.cr.Spec.VaultSecretManagement.WritePolicies {
		monitoringPolicyName := fmt.Sprintf("%s.%s-policy", r.monitoringProvider.GetServiceName(), r.cr.Namespace)
		monitoringPolicy, err := r.reconciler.ReadVaultPolicy(monitoringPolicyName)
		if err != nil {
			return err
		}
		if monitoringPolicy == "" || needToRefreshCredentials(monitoringSecret) {
			log.Info("Update policy for ZooKeeper Monitoring")
			monitoringPolicy := provider.BuildVaultPolicy(r.monitoringProvider.GetServiceName(), r.cr, "*")
			if err := r.reconciler.WriteVaultPolicy(monitoringPolicyName, monitoringPolicy); err != nil {
				return err
			}
		}
		zooKeeperClientPolicyName := fmt.Sprintf("%s.%s-client-policy", r.cr.Name, r.cr.Namespace)

		monitoringRoleName := fmt.Sprintf("%s.%s-role", r.monitoringProvider.GetServiceName(), r.cr.Namespace)
		monitoringRole, err := r.reconciler.ReadVaultAuthRole(monitoringRoleName, r.cr)
		if err != nil {
			return err
		}
		if monitoringRole == nil || needToRefreshCredentials(monitoringSecret) {
			log.Info("Update role for ZooKeeper Monitoring")
			monitoringRole := provider.BuildVaultRole(r.monitoringProvider.GetServiceAccountName(), r.cr, monitoringPolicyName, zooKeeperClientPolicyName)
			if err := r.reconciler.WriteVaultAuthRole(monitoringRoleName, monitoringRole, r.cr); err != nil {
				return err
			}
		}
	}

	if needToCleanSecret(monitoringSecret) {
		monitoringSecret.Annotations[refreshCredentialsAnnotation] = "false"
		if err := r.reconciler.cleanSecretData(monitoringSecret, log); err != nil {
			log.Error(err, "Cannot clean ZooKeeper Monitoring")
			return err
		}
		log.Info("ZooKeeper Monitoring secret was cleaned")
		r.reconciler.ResourceVersions[monitoringSecret.Name] = monitoringSecret.ResourceVersion
	}
	return nil
}
