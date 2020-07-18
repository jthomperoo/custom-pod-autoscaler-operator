/*
Copyright 2020 The Custom Pod Autoscaler Authors.

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

// Important: Run "make" and "make manifests" to regenerate code after modifying this file

import (
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomPodAutoscalerConfig defines the configuration options that can be passed to the CustomPodAutoscaler
type CustomPodAutoscalerConfig struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CustomPodAutoscalerSpec defines the desired state of CustomPodAutoscaler
type CustomPodAutoscalerSpec struct {
	// The image of the Custom Pod Autoscaler
	Template corev1.PodTemplateSpec `json:"template"`
	// ScaleTargetRef defining what the Custom Pod Autoscaler should manage
	ScaleTargetRef autoscaling.CrossVersionObjectReference `json:"scaleTargetRef"`
	// Configuration options to be delivered as environment variables to the container
	Config                  []CustomPodAutoscalerConfig `json:"config,omitempty"`
	ProvisionRole           *bool                       `json:"provisionRole,omitempty"`
	ProvisionRoleBinding    *bool                       `json:"provisionRoleBinding,omitempty"`
	ProvisionServiceAccount *bool                       `json:"provisionServiceAccount,omitempty"`
	ProvisionPod            *bool                       `json:"provisionPod,omitempty"`
}

// CustomPodAutoscalerStatus defines the observed state of CustomPodAutoscaler
type CustomPodAutoscalerStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cpa
// +groupName=custompodautoscaler.com
// CustomPodAutoscaler is the Schema for the custompodautoscalers API
type CustomPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPodAutoscalerSpec   `json:"spec,omitempty"`
	Status CustomPodAutoscalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomPodAutoscalerList contains a list of CustomPodAutoscaler
type CustomPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomPodAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomPodAutoscaler{}, &CustomPodAutoscalerList{})
}
