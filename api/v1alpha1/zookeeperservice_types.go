package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZooKeeper defines the specific ZooKeeper configuration
type ZooKeeper struct {
	DockerImage          string                  `json:"dockerImage"`
	Affinity             v1.Affinity             `json:"affinity,omitempty"`
	Replicas             int                     `json:"replicas"`
	Storage              Storage                 `json:"storage"`
	SnapshotStorage      SnapshotStorage         `json:"snapshotStorage,omitempty"`
	HeapSize             int                     `json:"heapSize"`
	Resources            v1.ResourceRequirements `json:"resources"`
	SecretName           string                  `json:"secretName"`
	QuorumAuthEnabled    bool                    `json:"quorumAuthEnabled,omitempty"`
	SecurityContext      v1.PodSecurityContext   `json:"securityContext,omitempty"`
	JolokiaPort          int32                   `json:"jolokiaPort,omitempty"`
	EnvironmentVariables []string                `json:"environmentVariables,omitempty"`
}

// Storage defines volumes of ZooKeeper
type Storage struct {
	Volumes   []string `json:"volumes,omitempty"`
	Nodes     []string `json:"nodes,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	ClassName []string `json:"className,omitempty"`
	Size      string   `json:"size"`
}

// Snapshot storage defines volume to store ZooKeeper snapshots
type SnapshotStorage struct {
	PersistentVolumeType      string `json:"persistentVolumeType,omitempty"`
	PersistentVolumeName      string `json:"persistentVolumeName,omitempty"`
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName,omitempty"`
	VolumeSize                string `json:"volumeSize"`
	NfsServer                 string `json:"nfsServer,omitempty"`
	NfsPath                   string `json:"nfsPath,omitempty"`
}

// Monitoring defines the specific ZooKeeper Monitoring configuration
type Monitoring struct {
	DockerImage               string                  `json:"dockerImage"`
	Affinity                  v1.Affinity             `json:"affinity,omitempty"`
	Resources                 v1.ResourceRequirements `json:"resources"`
	ZooKeeperHost             string                  `json:"zooKeeperHost"`
	ZooKeeperVolumes          string                  `json:"zooKeeperVolumes,omitempty"`
	NeedToCleanInfluxDb       bool                    `json:"needToCleanInfluxDb"`
	ZooKeeperBackupDaemonHost string                  `json:"zooKeeperBackupDaemonHost,omitempty"`
	SecretName                string                  `json:"secretName"`
	SmDbHost                  string                  `json:"smDbHost"`
	SmDbName                  string                  `json:"smDbName"`
	ZooKeeperJolokiaPort      int32                   `json:"zooKeeperJolokiaPort,omitempty"`
	SecurityContext           v1.PodSecurityContext   `json:"securityContext,omitempty"`
}

// BackupDaemon defines the specific ZooKeeper Backup Daemon configuration
type BackupDaemon struct {
	DockerImage     string                  `json:"dockerImage"`
	Affinity        v1.Affinity             `json:"affinity,omitempty"`
	BackupStorage   BackupStorage           `json:"backupStorage"`
	Resources       v1.ResourceRequirements `json:"resources"`
	BackupSchedule  string                  `json:"backupSchedule,omitempty"`
	EvictionPolicy  string                  `json:"evictionPolicy,omitempty"`
	IPv6            bool                    `json:"ipv6"`
	ZooKeeperHost   string                  `json:"zooKeeperHost"`
	ZooKeeperPort   int                     `json:"zooKeeperPort"`
	SecretName      string                  `json:"secretName"`
	SecurityContext v1.PodSecurityContext   `json:"securityContext,omitempty"`
}

// BackupStorage defines volume for ZooKeeper Backup Daemon
type BackupStorage struct {
	PersistentVolumeType      string  `json:"persistentVolumeType,omitempty"`
	PersistentVolumeName      string  `json:"persistentVolumeName,omitempty"`
	PersistentVolumeClaimName string  `json:"persistentVolumeClaimName,omitempty"`
	NodeName                  string  `json:"nodeName,omitempty"`
	VolumeSize                string  `json:"volumeSize"`
	PersistentVolumeLabel     string  `json:"persistentVolumeLabel,omitempty"`
	StorageClass              *string `json:"storageClass,omitempty"`
	NfsServer                 string  `json:"nfsServer,omitempty"`
	NfsPath                   string  `json:"nfsPath,omitempty"`
}

// ZooKeeperServiceSpec defines the desired state of ZooKeeperService
type ZooKeeperServiceSpec struct {
	ZooKeeper    *ZooKeeper    `json:"zooKeeper"`
	Monitoring   *Monitoring   `json:"monitoring,omitempty"`
	BackupDaemon *BackupDaemon `json:"backupDaemon,omitempty"`
}

// ZooKeeperServiceStatus defines the observed state of ZooKeeperService
type ZooKeeperServiceStatus struct {
	ZooKeeperStatus    ZooKeeperStatus    `json:"zooKeeperStatus,omitempty"`
	MonitoringStatus   MonitoringStatus   `json:"monitoringStatus,omitempty"`
	BackupDaemonStatus BackupDaemonStatus `json:"backupDaemonStatus,omitempty"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

func init() {
	SchemeBuilder.Register(&ZooKeeperService{}, &ZooKeeperServiceList{})
}
