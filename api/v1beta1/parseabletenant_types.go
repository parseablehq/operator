/*
 * Parseable Server (C) 2022 - 2023 Parseable, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ParseableTenantSpec defines the desired state of ParseableTenant
type ParseableTenantSpec struct {
	DeploymentOrder      []string                   `json:"deploymentOrder"`
	External             ExternalSpec               `json:"external,omitempty"`
	K8sConfigGroup       []K8sConfigGroupSpec       `json:"k8sConfigGroup"`
	ParseableConfigGroup []ParseableConfigGroupSpec `json:"parseableConfigGroup"`
	Metadata             Metadata                   `json:"metadata"`
	Nodes                []NodeSpec                 `json:"nodes"`
}

type ExternalSpec struct {
	ObjectStore ObjectStoreSpec `json:"objectStore"`
}

type ObjectStoreSpec struct {
	Spec ObjectStoreConfig `json:"spec"`
}

type ObjectStoreConfig struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type K8sConfigGroupSpec struct {
	Name               string            `json:"name"`
	Volumes            []v1.Volume       `json:"volumes,omitempty"`
	VolumeMount        []v1.VolumeMount  `json:"volumeMount,omitempty"`
	Image              string            `json:"image"`
	ImagePullPolicy    v1.PullPolicy     `json:"imagePullPolicy,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	Env                []v1.EnvVar       `json:"env,omitempty"`
	Tolerations        []v1.Toleration   `json:"tolerations,omitempty"`
	PodMetadata        Metadata          `json:"podMetadata,omitempty"`
	StorageConfig      []StorageConfig   `json:"storageConfig,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	Service            *v1.ServiceSpec   `json:"service,omitempty"`
}

type Metadata struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type StorageConfig struct {
	Name      string                       `json:"name"`
	MountPath string                       `json:"mountPath"`
	PvcSpec   v1.PersistentVolumeClaimSpec `json:"spec"`
}

type ParseableConfigGroupSpec struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type NodeSpec struct {
	Name                     string   `json:"name"`
	Kind                     string   `json:"kind"`
	NodeType                 string   `json:"nodeType"`
	Replicas                 int      `json:"replicas"`
	K8sConfigGroup           string   `json:"k8sConfigGroup"`
	CliArgs                  []string `json:"cliArgs"`
	ParseableConfigGroupName string   `json:"parseableConfigGroup"`
}

// ParseableTenantStatus defines the observed state of ParseableTenant
type ParseableTenantStatus struct {
	Version string `json:"statefulSets,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ParseableTenant is the Schema for the parseabletenants API
type ParseableTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ParseableTenantSpec   `json:"spec,omitempty"`
	Status ParseableTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ParseableTenantList contains a list of ParseableTenant
type ParseableTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ParseableTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ParseableTenant{}, &ParseableTenantList{})
}
