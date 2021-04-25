/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkflowRuntimeSpec defines the desired state of WorkflowRuntime
type WorkflowRuntimeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas defines the replication of the workflow runtime
	// Specificly, it determines the replication of Pods in its Deployment
	Replicas int32 `json:"replicas"`

	// Resources defines the resource provided to the Pod
	Resources corev1.ResourceRequirements `json:"resources"`

	// TODO: Add some fields
}

// WorkflowRuntimeStatus defines the observed state of WorkflowRuntime
type WorkflowRuntimeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Instances is a Pod List that WorkflowRuntime Manages
	Instances Instances `json:"instances"`
}

// Instances is a Pod List that WorkflowRuntime manages
// When the Deployment created or updated, Instances should be updated
// The key is the name of the Pod, for example "sample-c65c4f67-skbml"
type Instances map[string]Instance

// Instance records some runtime info of a Pod
// Specificly, it contains info about Function in the Pod and Pod metadata
type Instance struct {
	// Status describes metadata a Pod has
	Status InstanceStatus `json:"status"`
	// ProcessRuntimes is a list of ProcessRuntime
	ProcessRuntimes ProcessRuntimes `json:"processRuntimes"`
}

// InstanceStatus describes metadata a Pod has
type InstanceStatus struct {
	// CreationTimestamp is a timestamp representing the time when this Pod was created.
	CreationTimestamp metav1.Time `json:"creationTimestamp"`
	// IP address of the host to which the pod is assigned. Empty if not yet scheduled.
	HostIP string `json:"hostIP,omitempty"`
	// IP address allocated to the pod. Routable at least within the cluster. Empty if not yet allocated.
	PodIP string `json:"podIP,omitempty"`
}

// ProcessRuntimes is a list of ProcessRuntime
// The key is the name of the Function which is running in the Pod
type ProcessRuntimes map[string]ProcessRuntime

// ProcessRuntime records the process runtime info
type ProcessRuntime struct {
	// Number is the number the processes run the same Function
	Number int `json:"int"`
	// TODO: Add more fileds
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// WorkflowRuntime is the Schema for the workflowruntimes API
type WorkflowRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowRuntimeSpec   `json:"spec,omitempty"`
	Status WorkflowRuntimeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowRuntimeList contains a list of WorkflowRuntime
type WorkflowRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkflowRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkflowRuntime{}, &WorkflowRuntimeList{})
}
