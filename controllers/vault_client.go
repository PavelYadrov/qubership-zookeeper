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
	"encoding/json"
	"fmt"
	zookeeperservice "github.com/Netcracker/qubership-zookeeper/api/v1"
	"github.com/Netcracker/qubership-zookeeper/controllers/provider"
	"github.com/hashicorp/vault/api"
	kubeConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type VaultPasswordGenerator struct {
	passwordPolicyName string
	reconciler         *ZooKeeperServiceReconciler
}

var vaultClient *api.Client

func (r *ZooKeeperServiceReconciler) InitVaultClient(cr *zookeeperservice.ZooKeeperService) error {
	if err := checkVaultConnectionParameters(cr); err != nil {
		return err
	}
	config := &api.Config{
		Address: cr.Spec.VaultSecretManagement.Url,
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Error(err, "Error during creating vault client")
		return err
	}
	clientToken, err := r.login(cr)
	if err != nil {
		log.Error(err, "Error during login to vault")
		return err
	}
	client.SetToken(clientToken)
	vaultClient = client
	log.Info("Operator authenticated in vault using service account JWT")
	return nil
}

func (r *ZooKeeperServiceReconciler) login(cr *zookeeperservice.ZooKeeperService) (string, error) {
	kubernetesConfig := kubeConfig.GetConfigOrDie()
	jwtToken := kubernetesConfig.BearerToken

	options := map[string]interface{}{
		"jwt":  jwtToken,
		"role": cr.Spec.VaultSecretManagement.Role,
	}
	config := &api.Config{
		Address: cr.Spec.VaultSecretManagement.Url,
	}
	loginPath := "/auth/" + cr.Spec.VaultSecretManagement.Method + "/login"
	client, err := api.NewClient(config)
	if err != nil {
		log.Error(err, "Error occurred during creation Vault http client")
		return "", err
	}
	clientToken, err := client.Logical().Write(loginPath, options)
	if err != nil {
		log.Error(err, "Error occurred during authentication to vault with service account token")
		return "", err
	}

	return clientToken.Auth.ClientToken, nil
}

func checkVaultConnectionParameters(cr *zookeeperservice.ZooKeeperService) error {
	if cr.Spec.VaultSecretManagement.Url == "" {
		return fmt.Errorf("vault: Vault connection URL is empty")
	}
	if cr.Spec.VaultSecretManagement.Method == "" {
		return fmt.Errorf("vault: Vault authentication method is empty")
	}
	if cr.Spec.VaultSecretManagement.Role == "" {
		return fmt.Errorf("vault: Vault operator role is empty")
	}
	if cr.Spec.VaultSecretManagement.Path == "" {
		return fmt.Errorf("vault: Vault secret store path is empty")
	}
	return nil
}

func NewVaultPasswordGenerator(cr *zookeeperservice.ZooKeeperService, reconciler *ZooKeeperServiceReconciler) (*VaultPasswordGenerator, error) {
	passwordGenerationPolicyName := fmt.Sprintf("%s.%s-password-policy", cr.Name, cr.Namespace)
	passwordGenerationPolicy := provider.BuildVaultPasswordPolicy()
	if err := reconciler.WriteVaultPasswordPolicy(passwordGenerationPolicyName, passwordGenerationPolicy); err != nil {
		log.Error(err, "Cannot create vault password policy.")
		return nil, err
	}
	return &VaultPasswordGenerator{
		passwordPolicyName: passwordGenerationPolicyName,
		reconciler:         reconciler,
	}, nil
}

func (generator VaultPasswordGenerator) Generate() (string, error) {
	return generator.reconciler.GeneratePasswordForPolicy(generator.passwordPolicyName)
}

func (r *ZooKeeperServiceReconciler) WriteVaultSecret(path string, secretName string, secret map[string]interface{}) (int64, error) {
	secretPath := fmt.Sprintf("%s/data/%s", path, secretName)
	vaultSecret, err := vaultClient.Logical().Write(secretPath, map[string]interface{}{"data": secret})
	if err != nil {
		return 0, err
	}
	version, err := vaultSecret.Data["version"].(json.Number).Int64()
	log.Info(fmt.Sprintf("Secret '%s' was updated, new version is %d", secretPath, version))
	return version, err
}

func (r *ZooKeeperServiceReconciler) ReadVaultSecret(path string, secretName string) (map[string]interface{}, error) {
	secretPath := fmt.Sprintf("%s/data/%s", path, secretName)
	vaultSecret, err := vaultClient.Logical().Read(secretPath)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during loading secret '%s'", secretPath))
		return nil, err
	}
	if vaultSecret == nil {
		log.Info(fmt.Sprintf("Secret '%s' is not found", secretPath))
		return nil, nil
	}
	return vaultSecret.Data["data"].(map[string]interface{}), nil
}

func (r *ZooKeeperServiceReconciler) ReadVaultPolicy(policyName string) (string, error) {
	policy, err := vaultClient.Sys().GetPolicy(policyName)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during loading policy '%s'", policyName))
		return "", err
	}
	if policy == "" {
		log.Info(fmt.Sprintf("Policy '%s' is not found", policyName))
	}
	return policy, nil
}

func (r *ZooKeeperServiceReconciler) WriteVaultPolicy(policyName string, policy string) error {
	err := vaultClient.Sys().PutPolicy(policyName, policy)
	log.Info(fmt.Sprintf("Policy '%s' was updated", policyName))
	return err
}

func (r *ZooKeeperServiceReconciler) WriteVaultPasswordPolicy(policyName string, policy interface{}) error {
	request := vaultClient.NewRequest("PUT", fmt.Sprintf("/v1/sys/policies/password/%s", policyName))
	err := request.SetJSONBody(policy)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during buiding request for password policy '%s'", policyName))
		return err
	}
	if _, err := vaultClient.RawRequest(request); err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during writing password policy '%s'", policyName))
	}
	return err
}

func (r *ZooKeeperServiceReconciler) GeneratePasswordForPolicy(policyName string) (string, error) {
	request := vaultClient.NewRequest("GET", fmt.Sprintf("/v1/sys/policies/password/%s/generate", policyName))
	response, err := vaultClient.RawRequest(request)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during generate password for policy '%s'", policyName))
		return "", err
	}
	if response == nil {
		return "", fmt.Errorf("vault: Cannot generate password for policy: %s", policyName)
	}
	defer response.Body.Close()
	var respJson map[string]string
	if err = response.DecodeJSON(&respJson); err != nil {
		return "", err
	}
	return respJson["password"], nil
}

func (r *ZooKeeperServiceReconciler) WriteVaultAuthRole(roleName string, role interface{}, cr *zookeeperservice.ZooKeeperService) error {
	request := vaultClient.NewRequest("POST", fmt.Sprintf("/v1/auth/%s/role/%s", cr.Spec.VaultSecretManagement.Method, roleName))
	err := request.SetJSONBody(role)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during writing role '%s'", roleName))
		return err
	}
	_, err = vaultClient.RawRequest(request)
	return err
}

func (r *ZooKeeperServiceReconciler) ReadVaultAuthRole(roleName string, cr *zookeeperservice.ZooKeeperService) (map[string]interface{}, error) {
	request := vaultClient.NewRequest("GET", fmt.Sprintf("/v1/auth/%s/role/%s", cr.Spec.VaultSecretManagement.Method, roleName))
	response, err := vaultClient.RawRequest(request)
	if response != nil {
		defer response.Body.Close()
		if response.StatusCode == 404 {
			log.Info(fmt.Sprintf("Auth role '%s' is not found", roleName))
			return nil, nil
		}
	}
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred during reading role '%s'", roleName))
		return nil, err
	}

	var role map[string]interface{}
	err = response.DecodeJSON(&role)
	return role, err
}
