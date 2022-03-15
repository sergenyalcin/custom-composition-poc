/*
Copyright 2022.

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

type Function struct {
	// +kubebuilder:validation:Required
	Title string `json:"title"`

	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// +kubebuilder:validation:Optional
	Args string `json:"args,omitempty"`
}

// CustomCompositionSpec defines the desired state of CustomComposition
type CustomCompositionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	Resource string `json:"resource"`

	// +kubebuilder:validation:Required
	Functions []Function `json:"functions"`
}

type StatusCondition struct {
	Operation string `json:"operation,omitempty"`

	Message string `json:"message,omitempty"`

	Completed *bool `json:"completed,omitempty"`
}

// CustomCompositionStatus defines the observed state of CustomComposition
type CustomCompositionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State string `json:"state,omitempty"`

	Conditions []StatusCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomComposition is the Schema for the customcompositions API
type CustomComposition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomCompositionSpec   `json:"spec,omitempty"`
	Status CustomCompositionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomCompositionList contains a list of CustomComposition
type CustomCompositionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomComposition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomComposition{}, &CustomCompositionList{})
}
