/*
Copyright 2019 The Knative Authors.

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

	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Task is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
type Task struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the Task (from the client).
	// +optional
	Spec TaskSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the Task (from the controller).
	// +optional
	Status TaskStatus `json:"status,omitempty"`
}

// Check that Task can be validated and defaulted.
var _ apis.Validatable = (*Task)(nil)
var _ apis.Defaultable = (*Task)(nil)
var _ kmeta.OwnerRefable = (*Task)(nil)

// TaskSpec holds the desired state of the Task (from the client).
type TaskSpec struct {
	// Template describes the pods that will be created
	// +optional
	Template *corev1.PodTemplateSpec `json:"template,omitempty"`
}

const (
	// TaskConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	TaskConditionSucceeded = apis.ConditionSucceeded

	// TaskConditionAddressable has status true when this Task meets the
	// Addressable contract.
	TaskConditionAddressable apis.ConditionType = "Addressable"

	// TaskConditionResult tracks the job result.
	TaskConditionResult apis.ConditionType = "Result"
)

// TaskStatus communicates the observed state of the Task (from the controller).
type TaskStatus struct {
	duckv1beta1.Status `json:",inline"`

	// Address holds the information needed to connect this Addressable up to receive events.
	// +optional
	Address *duckv1beta1.Addressable `json:"address,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskList is a list of Task resources
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Task `json:"items"`
}
