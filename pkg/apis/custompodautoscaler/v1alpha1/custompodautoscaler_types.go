/*
Copyright 2019 The Custom Pod Autoscaler Authors.

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

// CustomPodAutoscalerConfig defines the configuration options that can be passed to the CustomPodAutoscaler
// +k8s:openapi-gen=true
type CustomPodAutoscalerConfig struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CustomPodAutoscalerSpec defines the desired state of CustomPodAutoscaler
// +k8s:openapi-gen=true
type CustomPodAutoscalerSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Image    string                      `json:"image"`
	Selector string                      `json:"selector,omitempty"`
	Config   []CustomPodAutoscalerConfig `json:"config,omitempty"`
}

// CustomPodAutoscalerStatus defines the observed state of CustomPodAutoscaler
// +k8s:openapi-gen=true
type CustomPodAutoscalerStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomPodAutoscaler is the Schema for the custompodautoscalers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cpa
// +groupName=custompodautoscaler.com
type CustomPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPodAutoscalerSpec   `json:"spec,omitempty"`
	Status CustomPodAutoscalerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomPodAutoscalerList contains a list of CustomPodAutoscaler
type CustomPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomPodAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomPodAutoscaler{}, &CustomPodAutoscalerList{})
}
