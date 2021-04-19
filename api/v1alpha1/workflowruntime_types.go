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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkflowRuntimeSpec defines the desired state of WorkflowRuntime
type WorkflowRuntimeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO: Add some fields
}

// WorkflowRuntimeStatus defines the observed state of WorkflowRuntime
type WorkflowRuntimeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replica defines the replication of the workflow runtime
	// Specificly, it determines the replication of Pods in its Deployment
	Replica int `json:"replica"`
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
