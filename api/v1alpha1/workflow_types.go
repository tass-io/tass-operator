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

// WorkflowSpec defines the desired state of Workflow
type WorkflowSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Environment represents the language environment of the code segments
	// The controller wil then launch the corresponding language environment
	// TODO: Is this field required? Discussion needed
	Environment Environment `json:"environment"`

	// Domain indicates the namespace of the workflow
	// A workflow can only use Functions in the same domain
	Domain string `json:"domain"`

	// Spec is a list of Flows
	Spec []Flow `json:"spec"`

	// TODO: Add more fields in the future
}

// Environment defines the language environments that tass supports
// +kubebuilder:validation:Enum=Golang;Python;JavaScript
type Environment string

const (
	// Golang means the language environment is Golang
	Golang Environment = "Golang"
	// Python means the language environment is Python
	Python Environment = "Python"
	// JavaScript means the language environment is JavaScript
	JavaScript Environment = "JavaScript"
)

// Flow defines the logic of a Function in a workflow
type Flow struct {
	// Name is the name of the flow which is unique in a workflow.
	// A function may be called multiple times in different places in a workflow.
	// So we need a Flow name to clear the logic.
	Name string `json:"name"`
	// Function is the function name which has been defined in Tass
	Function string `json:"function"`
	// Inputs specify which flows need to complete before this flow can start
	Inputs []string `json:"inputs"`
	// Outputs specify where the result of this flow should go
	Outputs []string `json:"outputs"`
	// Statement shows the flow control logic type
	// Valid values are:
	// - direct: The result of the flow go to downstream directly;
	// - switch: The result of the flow go to downstream based on the switch condition;
	// - loop: The result of the flow go back to itself until the loop condition break;
	Statement Statement `json:"statement"`
	// Condition is the control logic of the flow
	// Only worked when the Statement is 'Switch' or 'Loop'
	// +optional
	Condition Condition `json:"condition,omitempty"`
}

// Statement shows the flow control logic type
// +kubebuilder:validation:Enum=direct;switch;loop
type Statement string

const (
	// Direct is the result of the flow go to downstream directly
	Direct Statement = "direct"
	// Switch is the result of the flow go to downstream based on the switch condition;
	Switch Statement = "switch"
	// Loop is the result of the flow go back to itself until the loop condition break;
	Loop Statement = "loop"
)

// Condition is the control logic of the flow
type Condition struct {
	// Type is the data type that Tass workflow condition support
	// It also implicitly shows the result type of the flow
	// Valid values are:
	// - string: The condition type is string
	// - int: The condition type is int
	// - bool: The condition type is boolean
	Type ConditionType `json:"type"`
	// Operator defines the illegal operation in workflow condition statement
	// Valid values are:
	// - eq: The result is equal to the target
	// - ne: The result is not equal to the target
	// - lt: The result is less than the target
	// - le: The result is less than or equal to the target
	// - gt: The result is greater than the target
	// - ge: The result is greater than or equal to the target.
	Operator OperatorType `json:"operator"`
	// Comparision is used to compare with the flow result
	Comparision Comparision `json:"comparision"`
	// Destination defines the downstream Flows based on the condition result
	Destination Destination `json:"destination"`
}

// ConditionType is the data type that Tass workflow condition support
// +kubebuilder:validation:Enum=string;int;bool
type ConditionType string

const (
	// String means the condition type is string
	String ConditionType = "string"
	// Int means the condition type is int
	Int ConditionType = "int"
	// Bool means the condition type is boolean
	Bool ConditionType = "bool"
)

// OperatorType defines the illegal operation in workflow condition statement
// +kubebuilder:validation:Enum=eq;ne;lt;le;gt;ge
type OperatorType string

const (
	// Eq means the result is equal to the target
	Eq OperatorType = "eq"
	// Ne means the result is not equal to the target
	Ne OperatorType = "ne"
	// Lt means the result is less than the target, bool not accept
	Lt OperatorType = "lt"
	// Le means the result is less than or equal to the target, bool not accept
	Le OperatorType = "le"
	// Gt means the result is greater than the target, bool not accept
	Gt OperatorType = "gt"
	// Ge means the result is greater than or equal to the target, bool not accept
	Ge OperatorType = "ge"
)

// Comparision is used to compare with the flow result
// Comparision can be string, int or bool
// TODO: Validation needed
type Comparision string

// Destination defines the downstream Flows based on the condition result
type Destination struct {
	// IsTrue defines the downstream Flows if the condition is satisfied
	IsTrue []string `json:"isTrue"`
	// IsFalse defines the downstream Flows if the condition is satisfied
	IsFalse []string `json:"isFalse"`
}

// WorkflowStatus defines the observed state of Workflow
type WorkflowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status defined the observed state of workflow
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Workflow is the Schema for the workflows API
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflow
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
