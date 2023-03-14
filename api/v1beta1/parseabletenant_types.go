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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ParseableTenantSpec defines the desired state of ParseableTenant
type ParseableTenantSpec struct {
	DeploymentOrder      []string                   `json:"deploymentOrder"`
	External             ExternalSpec               `json:"external"`
	K8sConfigGroup       []K8sConfigGroupSpec       `json:"k8sConfigGroup"`
	ParseableConfigGroup []ParseableConfigGroupSpec `json:"parseableConfigGroup"`
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
	Name          string          `json:"name"`
	Volumes       []v1.Volume     `json:"volumes,omitempty"`
	Spec          v1.PodSpec      `json:"spec"`
	StorageConfig []StorageConfig `json:"storageConfig"`
}

type StorageConfig struct {
	Name    string                       `json:"name"`
	PvcSpec v1.PersistentVolumeClaimSpec `json:"spec"`
}

type ParseableConfigGroupSpec struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type NodeSpec struct {
	Name                 string `json:"name"`
	Kind                 string `json:"kind"`
	NodeType             string `json:"nodeType"`
	Replicas             int    `json:"replicas"`
	K8sConfigGroup       string `json:"k8sConfigGroup"`
	ParseableConfigGroup string `json:"parseableConfigGroup"`
}

// ParseableTenantStatus defines the observed state of ParseableTenant
type ParseableTenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
