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
	"strings"
)

const refreshCredentialsAnnotation = "vault.qubership.org/refresh-credentials"

type PasswordGenerator interface {
	Generate() (string, error)
}

func (r *ReconcileZooKeeper) processVaultSecrets(zooKeeperSecret *corev1.Secret) error {
	log.Info("Process vault secrets management for ZooKeeper")
	if r.cr.Status.VaultSecretManagementStatus.SecretVersions == nil {
		r.cr.Status.VaultSecretManagementStatus.SecretVersions = map[string]int{}
	}
	if r.cr.Spec.VaultSecretManagement.WritePolicies {
		zooKeeperAdminPolicyName := fmt.Sprintf("%s.%s-admin-policy", r.cr.Name, r.cr.Namespace)
		zooKeeperAdminPolicy, err := r.reconciler.ReadVaultPolicy(zooKeeperAdminPolicyName)
		if err != nil {
			return err
		}
		if zooKeeperAdminPolicy == "" || needToRefreshCredentials(zooKeeperSecret) {
			log.Info("Update admin policy for ZooKeeper")

			zooKeeperAdminPolicy := provider.BuildVaultPolicy(r.cr.Name, r.cr, "*")
			if err := r.reconciler.WriteVaultPolicy(zooKeeperAdminPolicyName, zooKeeperAdminPolicy); err != nil {
				return err
			}
		}

		zooKeeperAdminRoleName := fmt.Sprintf("%s.%s-role", r.cr.Name, r.cr.Namespace)
		zooKeeperAdminRole, err := r.reconciler.ReadVaultAuthRole(zooKeeperAdminRoleName, r.cr)
		if err != nil {
			return err
		}
		if zooKeeperAdminRole == nil || needToRefreshCredentials(zooKeeperSecret) {
			log.Info("Update admin role for ZooKeeper")

			zooKeeperAdminRole := provider.BuildVaultRole(r.zkProvider.GetServiceAccountName(), r.cr, zooKeeperAdminPolicyName)
			if err := r.reconciler.WriteVaultAuthRole(zooKeeperAdminRoleName, zooKeeperAdminRole, r.cr); err != nil {
				return err
			}
		}

		zooKeeperClientPolicyName := fmt.Sprintf("%s.%s-client-policy", r.cr.Name, r.cr.Namespace)
		zooKeeperClientPolicy, err := r.reconciler.ReadVaultPolicy(zooKeeperClientPolicyName)
		if err != nil {
			return err
		}
		if zooKeeperClientPolicy == "" || needToRefreshCredentials(zooKeeperSecret) {
			log.Info("Update client policy for ZooKeeper")

			zooKeeperClientPolicyRule := provider.BuildVaultPolicy(r.cr.Name, r.cr, "client-credentials")
			if err := r.reconciler.WriteVaultPolicy(zooKeeperClientPolicyName, zooKeeperClientPolicyRule); err != nil {
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

	if err := r.processAdminCredentials(zooKeeperSecret, passwordGenerator); err != nil {
		return err
	}

	if err := r.processClientCredentials(zooKeeperSecret, passwordGenerator); err != nil {
		return err
	}

	if err := r.processAdditionalUsersCredentials(zooKeeperSecret, passwordGenerator); err != nil {
		return err
	}
	if needToCleanSecret(zooKeeperSecret) {
		zooKeeperSecret.Annotations[refreshCredentialsAnnotation] = "false"
		if err := r.reconciler.cleanSecretData(zooKeeperSecret, log); err != nil {
			log.Error(err, "Cannot clean ZooKeeper secret")
			return err
		}
		log.Info("ZooKeeper secret was cleaned")
		r.reconciler.ResourceVersions[zooKeeperSecret.Name] = zooKeeperSecret.ResourceVersion
	}
	return nil
}

func (r *ReconcileZooKeeper) processAdminCredentials(zooKeeperSecret *corev1.Secret, passwordGenerator PasswordGenerator) error {
	adminCredentialsSecretName := fmt.Sprintf("%s.%s/admin-credentials", r.cr.Name, r.cr.Namespace)
	adminVaultSecret, err := r.reconciler.ReadVaultSecret(r.cr.Spec.VaultSecretManagement.Path, adminCredentialsSecretName)
	if err != nil {
		return err
	}
	if adminVaultSecret == nil || needToRefreshCredentials(zooKeeperSecret) {
		log.Info("Update admin credentials for ZooKeeper")
		var adminUsername string
		if secretContainsKey(zooKeeperSecret, "admin-username") {
			adminUsername = string(zooKeeperSecret.Data["admin-username"])
		} else {
			if adminVaultSecret != nil && adminVaultSecret["username"] != nil {
				adminUsername = adminVaultSecret["username"].(string)
			}

		}
		var adminPassword string
		if adminUsername != "" {
			adminPassword, err = passwordGenerator.Generate()
			if err != nil {
				return err
			}
		}
		adminCredentialsSecret := map[string]interface{}{
			"username": adminUsername,
			"password": adminPassword,
		}
		version, err := r.reconciler.WriteVaultSecret(r.cr.Spec.VaultSecretManagement.Path, adminCredentialsSecretName, adminCredentialsSecret)
		if err != nil {
			return err
		}
		r.cr.Status.VaultSecretManagementStatus.SecretVersions[adminCredentialsSecretName] = int(version)
	}
	return nil
}

func (r *ReconcileZooKeeper) processClientCredentials(zooKeeperSecret *corev1.Secret, passwordGenerator PasswordGenerator) error {
	clientCredentialsSecretName := fmt.Sprintf("%s.%s/client-credentials", r.cr.Name, r.cr.Namespace)
	clientVaultSecret, err := r.reconciler.ReadVaultSecret(r.cr.Spec.VaultSecretManagement.Path, clientCredentialsSecretName)
	if err != nil {
		return err
	}
	if clientVaultSecret == nil || needToRefreshCredentials(zooKeeperSecret) {
		log.Info("Update client credentials for ZooKeeper")

		var clientUsername string
		if secretContainsKey(zooKeeperSecret, "client-username") {
			clientUsername = string(zooKeeperSecret.Data["client-username"])
		} else {
			if clientVaultSecret != nil && clientVaultSecret["username"] != nil {
				clientUsername = clientVaultSecret["username"].(string)
			}
		}
		var clientPassword string
		if clientUsername != "" {
			clientPassword, err = passwordGenerator.Generate()
			if err != nil {
				return err
			}
		}
		clientCredentialsSecret := map[string]interface{}{
			"username": clientUsername,
			"password": clientPassword,
		}
		version, err := r.reconciler.WriteVaultSecret(r.cr.Spec.VaultSecretManagement.Path, clientCredentialsSecretName, clientCredentialsSecret)
		if err != nil {
			return err
		}
		r.cr.Status.VaultSecretManagementStatus.SecretVersions[clientCredentialsSecretName] = int(version)
	}
	return nil
}

func (r *ReconcileZooKeeper) processAdditionalUsersCredentials(zooKeeperSecret *corev1.Secret, passwordGenerator PasswordGenerator) error {
	additionalUsersSecretName := fmt.Sprintf("%s.%s/additional-users", r.cr.Name, r.cr.Namespace)
	additionalUsersSecret, err := r.reconciler.ReadVaultSecret(r.cr.Spec.VaultSecretManagement.Path, additionalUsersSecretName)
	if err != nil {
		return err
	}
	if additionalUsersSecret == nil || needToRefreshCredentials(zooKeeperSecret) {
		log.Info("Update additional users credentials for ZooKeeper")
		var userNames []string
		if secretContainsKey(zooKeeperSecret, "additional-users") {
			userNames = r.extractUserNames(string(zooKeeperSecret.Data["additional-users"]))
		} else {
			if additionalUsersSecret != nil && additionalUsersSecret["users"] != nil {
				userNames = r.extractUserNames(additionalUsersSecret["users"].(string))
			}
		}
		var additionalUsers []string
		for _, username := range userNames {
			if username != "" {
				pass, err := passwordGenerator.Generate()
				if err != nil {
					return err
				}
				additionalUsers = append(additionalUsers, fmt.Sprintf("%s:%s", username, pass))
			}
		}

		additionalUsersSecret := map[string]interface{}{
			"users": strings.Join(additionalUsers, ","),
		}
		version, err := r.reconciler.WriteVaultSecret(r.cr.Spec.VaultSecretManagement.Path, additionalUsersSecretName, additionalUsersSecret)
		if err != nil {
			return err
		}
		r.cr.Status.VaultSecretManagementStatus.SecretVersions[additionalUsersSecretName] = int(version)
	}
	return nil
}

func (r *ReconcileZooKeeper) extractUserNames(userNamesString string) []string {
	userNames := strings.Split(userNamesString, ",")
	for k, v := range userNames {
		username := strings.Split(v, ":")[0]
		userNames[k] = username
	}
	return userNames
}

func needToRefreshCredentials(zooKeeperSecret *corev1.Secret) bool {
	return zooKeeperSecret.Annotations != nil && zooKeeperSecret.Annotations[refreshCredentialsAnnotation] == "true"
}

func needToCleanSecret(secret *corev1.Secret) bool {
	return secret.Name != "" && (secret.Data != nil || needToRefreshCredentials(secret))
}
