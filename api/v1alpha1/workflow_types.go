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

	// Env is the environment variables for the Workflow
	// It is defined by users
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// Spec is a list of Flows
	Spec []Flow `json:"spec"`

	// TODO: Add more fields in the future
}

// Flow defines the logic of a Function in a workflow
type Flow struct {
	// Name is the name of the flow which is unique in a workflow.
	// A function may be called multiple times in different places in a workflow.
	// So we need a Flow name to clear the logic.
	Name string `json:"name"`
	// Function is the function name which has been defined in Tass
	Function string `json:"function"`
	// Outputs specify where the result of this flow should go
	// +optional
	Outputs []string `json:"outputs"`
	// Statement shows the flow control logic type
	// Valid values are:
	// - direct: The result of the flow go to downstream directly;
	// - switch: The result of the flow go to downstream based on the switch condition;
	Statement Statement `json:"statement"`
	// Role is the role of the Flow
	// Valid values are:
	// - start: The role of the Flow is "start" which means it is the entrance of workflow instance
	// - end: The role of the Flow is "end" which means it is the exit point of workflow instance
	// - orphan: The role of the Flow is "orphan" which is a special case that
	// the workflow instance has only one function
	// If no value is specified, it means this is an intermediate Flow instance
	// +optional
	Role Role `json:"role,omitempty"`

	// Conditions are the control logic group of the flow
	// The first element of the Conditions is the root control logic
	// Only worked when the Statement is 'Switch'
	// +optional
	Conditions []*Condition `json:"conditions,omitempty"`
}

// Statement shows the flow control logic type
// +kubebuilder:validation:Enum=direct;switch
type Statement string

const (
	// Direct is the result of the flow go to downstream directly
	Direct Statement = "direct"
	// Switch is the result of the flow go to downstream based on the switch condition;
	Switch Statement = "switch"
)

// Role is the role of the Flow
// +kubebuilder:validation:Enum=start;end
type Role string

const (
	// Start means the role of the Flow is "start" which means it is the entrance of workflow instance
	Start Role = "start"
	// End means the role of the Flow is "end" which means it is the exit point of workflow instance
	End Role = "end"
	// Orphan means the role of the Flow is "orphan" which is a special case that
	// the workflow instance has only one function
	Orphan Role = "orphan"
)

// Condition is the control logic of the flow
// A sample of Condition
// ```yaml
// condition:
// 	 name: root
// 	 type: int
// 	 operator: gt
// 	 target: $.a
// 	 comparision: 50
// 	 destination:
// 		 isTrue:  # ...
// 		 isFalse: # ...
// ```
// It is same as:
// if $.a >= 50 {
// 	 goto isTrue logic
// } else {
// 	 goto isFalse logic
// }
type Condition struct {
	// Name is the name of a Condition, it's unique in a Condition group
	Name string `json:"name"`
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
	// Target shows the specific data that the flow result uses to compare with
	// The result of the flow can be a simple type like string, bool or int
	// But it can also be a complex object contains some fileds
	// Whatever the result is, the Flow runtime will wrap the result to a JSON object
	// to unifiy the transmission process.
	// For example, the result of the user code is a string type, let's say "tass",
	// and then it will be wrapped as a JSON object {"$": "tass"} as the result of the Flow.
	//
	// If the result of user code is not a simple type, it can be much more complex
	// For example, the Flow result can be {"$":{"name": "tass","type": "faas"}}
	// So in this case, if we want to use the "type" property
	// to compare with Comparision, the Target value should be "$.type"
	//
	// If users don't specify the Target field,or the Target value is just "$",
	// it means the user code result is just a simple type
	// Otherwise, the user must provide a Target value to claim the property to use
	// One more example to show how to get the key in Flow result
	// Let's say the result is {"$":{"name":"tass","info":{"type":"fn","timeout":60}}}
	// We want the "timeout" key, so the Target value is "$.info.timeout"
	Target string `json:"target,omitempty"`
	// Comparision is used to compare with the flow result
	// Comparision can be a realistic value, like "cash", "5", "true"
	// it can also be a property of the flow result, like "$.b"
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
// When a Flow finishes its task, the result of the Flow goes to the downstream Flows
// After passing a Condition, the Flows where result goes is determinated by Destination field
// The result can go to the downstream Flows directly,
// or it needs a new round of Conditions, or both.
// Here is a sample of Destination:
// ```yaml
// destination:
//	 isTrue:
//	 	 flows:
//	 	 - flow-a       # this is a Flow Name
//		 - flow-b
//	 isFalse:
//	 	 conditions:
//	   - condition-a  # this is a Condition name
// ``
type Destination struct {
	// IsTrue defines the downstream Flows if the condition is satisfied
	IsTrue Next `json:"isTrue"`
	// IsFalse defines the downstream Flows if the condition is not satisfied
	IsFalse Next `json:"isFalse"`
}

// Next shows the next Condition or Flows the data goes
type Next struct {
	// Flows lists the Flows where the result of the current Flow goes
	Flows []string `json:"flows,omitempty"`
	// Condition lists the Condition where the result of the current Flow goes
	// It means that the result needs more control logic check
	Conditions []*Condition `json:"conditions,omitempty"`
}

// WorkflowStatus defines the observed state of Workflow
type WorkflowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
