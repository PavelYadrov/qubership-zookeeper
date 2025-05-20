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

package provider

import (
	"fmt"
	zookeeperservice "github.com/Netcracker/qubership-zookeeper/api/v1"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"strings"
)

func IsVaultSecretManagementEnabled(cr *zookeeperservice.ZooKeeperService) bool {
	return cr.Spec.VaultSecretManagement != nil && cr.Spec.VaultSecretManagement.Enabled
}

// getVaultSecretEnvVarSource
func getVaultSecretEnvVarSource(serviceName string, cr *zookeeperservice.ZooKeeperService, secret string, key string) string {
	return fmt.Sprintf("vault:/%s/data/%s.%s/%s#%s", cr.Spec.VaultSecretManagement.Path, serviceName, cr.Namespace, secret, key)
}

func getVaultConnectionEnvVars(serviceName string, cr *zookeeperservice.ZooKeeperService) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "VAULT_SKIP_VERIFY", Value: "True"},
		{Name: "VAULT_ADDR", Value: cr.Spec.VaultSecretManagement.Url},
		{Name: "VAULT_PATH", Value: cr.Spec.VaultSecretManagement.Method},
		{Name: "VAULT_ROLE", Value: fmt.Sprintf("%s.%s-role", serviceName, cr.Namespace)},
		{Name: "VAULT_IGNORE_MISSING_SECRETS", Value: "False"},
	}
}

func getVaultInitContainer(cr *zookeeperservice.ZooKeeperService) corev1.Container {
	return corev1.Container{
		Name:            "copy-vault-env",
		Image:           cr.Spec.VaultSecretManagement.DockerImage,
		Command:         []string{"sh", "-c", "cp /usr/local/bin/vault-env /vault/"},
		VolumeMounts:    []corev1.VolumeMount{{Name: "vault-env", MountPath: "/vault"}},
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: getDefaultContainerSecurityContext(),
	}
}

func BuildVaultRole(serviceAccountName string, cr *zookeeperservice.ZooKeeperService, policyNames ...string) interface{} {
	return map[string]interface{}{
		"bound_service_account_names":      serviceAccountName,
		"bound_service_account_namespaces": cr.Namespace,
		"policies":                         policyNames,
		"ttl":                              "30m",
	}
}

func BuildVaultPolicy(serviceName string, cr *zookeeperservice.ZooKeeperService, secretNamePatterns ...string) string {
	var policies []string
	for _, secretNamePattern := range secretNamePatterns {
		policies = append(policies, fmt.Sprintf("path \"%s/data/%s.%s/%s\" \n{\n\tcapabilities = [\"read\", \"list\"]\n}", cr.Spec.VaultSecretManagement.Path, serviceName, cr.Namespace, secretNamePattern))
	}
	return strings.Join(policies, "\n")
}

func BuildVaultPolicyForPath(paths ...string) string {
	var policies []string
	for _, secretNamePattern := range paths {
		policies = append(policies, fmt.Sprintf("path \"%s\" \n{\n\tcapabilities = [\"read\", \"list\"]\n}", secretNamePattern))
	}
	return strings.Join(policies, "\n")
}

func BuildVaultPasswordPolicy() interface{} {
	return map[string]interface{}{
		"policy": "length = 10" +
			"rule \"charset\" {" +
			"	charset = \"abcdefghijklmnopqrstuvwxyz\"" +
			"	min-chars = 3" +
			"}" +
			"rule \"charset\" {" +
			"	charset = \"ABCDEFGHIJKLMNOPQRSTUVWXYZ\"" +
			"	min-chars = 3" +
			"}" +
			"rule \"charset\" {" +
			"		charset = \"0123456789\"" +
			"	min-chars = 1" +
			"}" +
			"rule \"charset\" {" +
			"	charset = \"_!" +
			"	min-chars = 1" +
			"}",
	}
}

func GetVaultSecretForPath(secretPathsForComponent map[string]string, key string) string {
	if secretPathsForComponent != nil {
		path := secretPathsForComponent[key]
		if strings.HasPrefix(path, "vault:/") {
			var re = regexp.MustCompile("(vault:/)(.*)(#.*)")
			return re.ReplaceAllString(path, `$2`)
		}
	}
	return ""
}
