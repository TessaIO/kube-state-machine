/*
Copyright 2024.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StateType defines the type of a given state
type StateType string

const (
	// StateTypeChoice is a state that can have multiple possible outcomes
	StateTypeChoice StateType = "Choice"
	// StateTypePass is a state that does nothing
	StateTypePass StateType = "Pass"
	// StateTypeWait is a state that waits for a specified time, or a valid condition
	StateTypeWait StateType = "Wait"
	// StateTypeSucceed is a state that terminates a state machine successfully
	StateTypeSucceed StateType = "Succeed"
	// StateTypeFail is a state that terminates a state machine with a failure
	StateTypeFail StateType = "Fail"
	// StateTypeParallel is a state that can have multiple branches of execution
	StateTypeParallel StateType = "Parallel"
	// StateTypeTask is a state that can be used to deploy a given workload
	StateTypeTask StateType = "Task"
)

// State defines the single state of a state machine
type Condition struct {
	PodName string `json:"podName"`
	Status  string `json:"status"`
}

// State defines the single state of a state machine
type Choice struct {
	Name      string    `json:"name"`
	Condition Condition `json:"condition"`
	Next      string    `json:"next"`
}

// State defines the single state of a state machine
type State struct {
	Name    string          `json:"name"`
	Type    StateType       `json:"type"`
	Choices *[]Choice       `json:"choices,omitempty"`
	Task    *corev1.PodSpec `json:"task,omitempty"`
	End     *bool           `json:"end,omitempty"`
	Next    *string         `json:"next,omitempty"`
}

// StateMachineSpec defines the desired state of StateMachine
type StateMachineSpec struct {
	States []State `json:"states,omitempty"`
}

// StateMachineStatus defines the observed state of StateMachine
type StateMachineStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StateMachine is the Schema for the statemachines API
type StateMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StateMachineSpec   `json:"spec,omitempty"`
	Status StateMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StateMachineList contains a list of StateMachine
type StateMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StateMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StateMachine{}, &StateMachineList{})
}
