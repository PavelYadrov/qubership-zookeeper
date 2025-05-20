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
	"github.com/Netcracker/qubership-zookeeper/util"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcileBackupDaemon) processVaultSecrets(backupDaemonSecret *corev1.Secret) error {
	log.Info("Process vault secrets management for ZooKeeper Backup Daemon")
	if r.cr.Spec.VaultSecretManagement.WritePolicies {
		backupDaemonPolicyName := fmt.Sprintf("%s.%s-policy", r.backupDaemonProvider.GetServiceName(), r.cr.Namespace)
		backupDaemonPolicy, err := r.reconciler.ReadVaultPolicy(backupDaemonPolicyName)
		if err != nil {
			return err
		}
		if backupDaemonPolicy == "" || needToRefreshCredentials(backupDaemonSecret) {
			log.Info("Update policy for ZooKeeper Backup Daemon")

			backupDaemonPolicy := provider.BuildVaultPolicy(r.backupDaemonProvider.GetServiceName(), r.cr, "*")
			if err := r.reconciler.WriteVaultPolicy(backupDaemonPolicyName, backupDaemonPolicy); err != nil {
				return err
			}
		}
		zooKeeperAdminPolicyName := fmt.Sprintf("%s.%s-admin-policy", r.cr.Name, r.cr.Namespace)

		backupDaemonRoleName := fmt.Sprintf("%s.%s-role", r.backupDaemonProvider.GetServiceName(), r.cr.Namespace)
		backupDaemonRole, err := r.reconciler.ReadVaultAuthRole(backupDaemonRoleName, r.cr)
		if err != nil {
			return err
		}
		if backupDaemonRole == nil || needToRefreshCredentials(backupDaemonSecret) {
			log.Info("Update role for ZooKeeper Backup Daemon")

			monitoringRole := provider.BuildVaultRole(r.backupDaemonProvider.GetServiceAccountName(), r.cr, backupDaemonPolicyName, zooKeeperAdminPolicyName)
			if err := r.reconciler.WriteVaultAuthRole(backupDaemonRoleName, monitoringRole, r.cr); err != nil {
				return err
			}
		}
	}

	var passwordGenerator PasswordGenerator
	var err error
	if r.cr.Spec.VaultSecretManagement.PasswordGenerationMechanism == "vault" {
		passwordGenerator, err = NewVaultPasswordGenerator(r.cr, r.reconciler)
		if err != nil {
			log.Error(err, "Cannot create vault password generator. Try to use `operator` password generation mechanism")
			return err
		}
	} else {
		passwordGenerator, err = util.NewOperatorPasswordGenerator()
		if err != nil {
			log.Error(err, "Cannot create operator password generator.")
			return err
		}
	}

	if err := r.processCredentials(backupDaemonSecret, passwordGenerator); err != nil {
		return err
	}

	if needToCleanSecret(backupDaemonSecret) {
		backupDaemonSecret.Annotations[refreshCredentialsAnnotation] = "false"
		if err := r.reconciler.cleanSecretData(backupDaemonSecret, log); err != nil {
			log.Error(err, "Cannot clean ZooKeeper Backup Daemon secret")
			return err
		}
		log.Info("ZooKeeper Backup Daemon secret was cleaned")
		r.reconciler.ResourceVersions[backupDaemonSecret.Name] = backupDaemonSecret.ResourceVersion
	}
	return nil
}

func (r *ReconcileBackupDaemon) processCredentials(backupDaemonSecret *corev1.Secret, passwordGenerator PasswordGenerator) error {
	credentialsSecretName := fmt.Sprintf("%s.%s/credentials", r.backupDaemonProvider.GetServiceName(), r.cr.Namespace)
	vaultSecret, err := r.reconciler.ReadVaultSecret(r.cr.Spec.VaultSecretManagement.Path, credentialsSecretName)
	if err != nil {
		return err
	}
	if vaultSecret == nil || needToRefreshCredentials(backupDaemonSecret) {
		log.Info("Update credentials for ZooKeeper Backup Daemon")
		var username string
		if secretContainsKey(backupDaemonSecret, "username") {
			username = string(backupDaemonSecret.Data["username"])
		} else {
			if vaultSecret != nil && vaultSecret["username"] != nil {
				username = vaultSecret["username"].(string)
			}

		}
		var password string
		if username != "" {
			password, err = passwordGenerator.Generate()
			if err != nil {
				return err
			}
		}
		credentialsSecret := map[string]interface{}{
			"username": username,
			"password": password,
		}
		version, err := r.reconciler.WriteVaultSecret(r.cr.Spec.VaultSecretManagement.Path, credentialsSecretName, credentialsSecret)
		if err != nil {
			return err
		}
		r.cr.Status.VaultSecretManagementStatus.SecretVersions[credentialsSecretName] = int(version)
	}
	return nil
}
