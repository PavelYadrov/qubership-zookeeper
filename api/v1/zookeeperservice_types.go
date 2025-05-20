package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZooKeeper defines the specific ZooKeeper configuration
type ZooKeeper struct {
	DockerImage          string                  `json:"dockerImage"`
	Affinity             v1.Affinity             `json:"affinity,omitempty"`
	Tolerations          []v1.Toleration         `json:"tolerations,omitempty"`
	PriorityClassName    string                  `json:"priorityClassName,omitempty"`
	Replicas             int                     `json:"replicas"`
	Storage              Storage                 `json:"storage"`
	SnapshotStorage      SnapshotStorage         `json:"snapshotStorage,omitempty"`
	HeapSize             int                     `json:"heapSize"`
	Resources            v1.ResourceRequirements `json:"resources"`
	SecretName           string                  `json:"secretName"`
	QuorumAuthEnabled    bool                    `json:"quorumAuthEnabled,omitempty"`
	Ssl                  Ssl                     `json:"ssl,omitempty"`
	SecurityContext      v1.PodSecurityContext   `json:"securityContext,omitempty"`
	JolokiaPort          int32                   `json:"jolokiaPort,omitempty"`
	EnvironmentVariables []string                `json:"environmentVariables,omitempty"`
	RollingUpdate        bool                    `json:"rollingUpdate,omitempty"`
	CustomLabels         map[string]string       `json:"customLabels,omitempty"`
	Diagnostics          Diagnostics             `json:"diagnostics,omitempty"`
	AuditEnabled         bool                    `json:"auditEnabled,omitempty"`
}

// Storage defines volumes of ZooKeeper
type Storage struct {
	Volumes   []string `json:"volumes,omitempty"`
	Nodes     []string `json:"nodes,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	ClassName []string `json:"className,omitempty"`
	Size      string   `json:"size"`
}

// SnapshotStorage defines volume to store ZooKeeper snapshots
type SnapshotStorage struct {
	PersistentVolumeType      string  `json:"persistentVolumeType,omitempty"`
	PersistentVolumeName      string  `json:"persistentVolumeName,omitempty"`
	PersistentVolumeClaimName string  `json:"persistentVolumeClaimName,omitempty"`
	NodeName                  string  `json:"nodeName,omitempty"`
	VolumeSize                string  `json:"volumeSize"`
	PersistentVolumeLabel     string  `json:"persistentVolumeLabel,omitempty"`
	StorageClass              *string `json:"storageClass,omitempty"`
	// Deprecated: NFS persistent volume creation is not supported, this is for backward compatibility
	NfsServer string `json:"nfsServer,omitempty"`
	// Deprecated: NFS persistent volume creation is not supported, this is for backward compatibility
	NfsPath string `json:"nfsPath,omitempty"`
}

// Ssl defines Ssl ZooKeeper settings
type Ssl struct {
	CipherSuites            []string `json:"cipherSuites,omitempty"`
	EnableTwoWaySsl         bool     `json:"enableTwoWaySsl,omitempty"`
	AllowNonencryptedAccess bool     `json:"allowNonencryptedAccess,omitempty"`
}

// S3 defines parameters for S3 storage for ZooKeeper Backup Daemon
type S3 struct {
	Enabled       bool   `json:"enabled,omitempty"`
	Url           string `json:"url,omitempty"`
	Bucket        string `json:"bucket,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
	SslVerify     bool   `json:"sslVerify,omitempty"`
	SslSecretName string `json:"sslSecretName,omitempty"`
	SslCert       string `json:"sslCert,omitempty"`
}

// Monitoring defines the specific ZooKeeper Monitoring configuration
type Monitoring struct {
	DockerImage               string                  `json:"dockerImage"`
	Affinity                  v1.Affinity             `json:"affinity,omitempty"`
	Tolerations               []v1.Toleration         `json:"tolerations,omitempty"`
	PriorityClassName         string                  `json:"priorityClassName,omitempty"`
	Resources                 v1.ResourceRequirements `json:"resources"`
	MonitoringType            string                  `json:"monitoringType"`
	ZooKeeperHost             string                  `json:"zooKeeperHost"`
	ZooKeeperBackupDaemonHost string                  `json:"zooKeeperBackupDaemonHost,omitempty"`
	SecretName                string                  `json:"secretName"`
	ZooKeeperJolokiaPort      int32                   `json:"zooKeeperJolokiaPort,omitempty"`
	SecurityContext           v1.PodSecurityContext   `json:"securityContext,omitempty"`
	CustomLabels              map[string]string       `json:"customLabels,omitempty"`
	// Deprecated: Influx DB is no longer supported, this is for backward compatibility
	ZooKeeperVolumes string `json:"zooKeeperVolumes,omitempty"`
	// Deprecated: Influx DB is no longer supported, this is for backward compatibility
	NeedToCleanInfluxDb bool `json:"needToCleanInfluxDb,omitempty"`
	// Deprecated: Influx DB is no longer supported, this is for backward compatibility
	SmDbHost string `json:"smDbHost,omitempty"`
	// Deprecated: Influx DB is no longer supported, this is for backward compatibility
	SmDbName string `json:"smDbName,omitempty"`
}

// VaultSecretManagement defines Vault secret management configuration
type VaultSecretManagement struct {
	DockerImage                 string      `json:"dockerImage"`
	Enabled                     bool        `json:"enabled,omitempty"`
	Path                        string      `json:"path,omitempty"`
	Url                         string      `json:"url,omitempty"`
	Role                        string      `json:"role,omitempty"`
	Method                      string      `json:"method,omitempty"`
	PasswordGenerationMechanism string      `json:"passwordGenerationMechanism,omitempty"`
	WritePolicies               bool        `json:"writePolicies,omitempty"`
	SecretPaths                 SecretPaths `json:"secretPaths,omitempty"`
}

type SecretPaths struct {
	Monitoring map[string]string `json:"monitoring,omitempty"`
}

type Global struct {
	WaitForPodsReady bool              `json:"waitForPodsReady"`
	PodsReadyTimeout int               `json:"podReadinessTimeout"`
	CustomLabels     map[string]string `json:"customLabels,omitempty"`
	DefaultLabels    map[string]string `json:"defaultLabels,omitempty"`
	ZooKeeperSsl     ZooKeeperSsl      `json:"zooKeeperSsl,omitempty"`
}

// ZooKeeperSsl shows ssl configuration
type ZooKeeperSsl struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secretName,omitempty"`
}

// BackupDaemonSsl shows ssl configuration
type BackupDaemonSsl struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secretName,omitempty"`
}

// IntegrationTests defines integration tests configuration
type IntegrationTests struct {
	ServiceName      string `json:"serviceName"`
	WaitForResult    bool   `json:"waitForResult"`
	Timeout          int    `json:"timeout"`
	RandomRunTrigger string `json:"randomRunTrigger,omitempty"`
}

// BackupDaemon defines the specific ZooKeeper Backup Daemon configuration
type BackupDaemon struct {
	DockerImage       string                  `json:"dockerImage"`
	Affinity          v1.Affinity             `json:"affinity,omitempty"`
	Tolerations       []v1.Toleration         `json:"tolerations,omitempty"`
	PriorityClassName string                  `json:"priorityClassName,omitempty"`
	BackupStorage     SnapshotStorage         `json:"backupStorage"`
	Resources         v1.ResourceRequirements `json:"resources"`
	BackupSchedule    string                  `json:"backupSchedule,omitempty"`
	S3                *S3                     `json:"s3,omitempty"`
	EvictionPolicy    string                  `json:"evictionPolicy,omitempty"`
	IPv6              bool                    `json:"ipv6"`
	ZooKeeperHost     string                  `json:"zooKeeperHost"`
	ZooKeeperPort     int                     `json:"zooKeeperPort"`
	SecretName        string                  `json:"secretName"`
	SecurityContext   v1.PodSecurityContext   `json:"securityContext,omitempty"`
	CustomLabels      map[string]string       `json:"customLabels,omitempty"`
	BackupDaemonSsl   BackupDaemonSsl         `json:"backupDaemonSsl,omitempty"`
}

// ZooKeeperServiceSpec defines the desired state of ZooKeeperService
type ZooKeeperServiceSpec struct {
	Global                *Global                `json:"global,omitempty"`
	ZooKeeper             *ZooKeeper             `json:"zooKeeper"`
	Monitoring            *Monitoring            `json:"monitoring,omitempty"`
	BackupDaemon          *BackupDaemon          `json:"backupDaemon,omitempty"`
	VaultSecretManagement *VaultSecretManagement `json:"vaultSecretManagement,omitempty"`
	IntegrationTests      *IntegrationTests      `json:"integrationTests,omitempty"`
}

// ZooKeeperServiceStatus defines the observed state of ZooKeeperService
type ZooKeeperServiceStatus struct {
	ZooKeeperStatus             ZooKeeperStatus             `json:"zooKeeperStatus,omitempty"`
	MonitoringStatus            MonitoringStatus            `json:"monitoringStatus,omitempty"`
	BackupDaemonStatus          BackupDaemonStatus          `json:"backupDaemonStatus,omitempty"`
	VaultSecretManagementStatus VaultSecretManagementStatus `json:"vaultSecretManagementStatus,omitempty"`
	Conditions                  []StatusCondition           `json:"conditions,omitempty"`
}

type ZooKeeperStatus struct {
	Servers []string `json:"servers,omitempty"`
}

type MonitoringStatus struct {
	Nodes []string `json:"nodes,omitempty"`
}

type BackupDaemonStatus struct {
	Nodes []string `json:"nodes,omitempty"`
}

type VaultSecretManagementStatus struct {
	SecretVersions map[string]int `json:"secretVersions,omitempty"`
}

// StatusCondition contains description of status of ZooKeeperService
type StatusCondition struct {
	// Type - Can be "In progress", "Failed", "Successful" or "Ready".
	Type string `json:"type"`
	// Status - "True" if condition is successfully done and "False" if condition has failed or in progress type.
	Status string `json:"status"`
	// Reason - One-word CamelCase reason for the condition's state.
	Reason string `json:"reason"`
	// Message - Human-readable message indicating details about last transition.
	Message string `json:"message"`
	// LastTransitionTime - Last time the condition transit from one status to another.
	LastTransitionTime string `json:"lastTransitionTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// ZooKeeperService is the Schema for the zookeeperservices API
type ZooKeeperService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZooKeeperServiceSpec   `json:"spec,omitempty"`
	Status ZooKeeperServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ZooKeeperServiceList contains a list of ZooKeeperService
type ZooKeeperServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZooKeeperService `json:"items"`
}

// Diagnostics defines diagnostic configuration
type Diagnostics struct {
	// +kubebuilder:validation:Enum=disable;off;dev;prod
	// +kubebuilder:default=disable
	Mode         string `json:"mode,omitempty"`
	AgentService string `json:"agentService,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ZooKeeperService{}, &ZooKeeperServiceList{})
}
