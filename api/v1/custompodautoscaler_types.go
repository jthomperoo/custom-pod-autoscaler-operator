/*
Copyright 2022 The Custom Pod Autoscaler Authors.

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

// Important: Run "make generate" to regenerate code after modifying this file

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	autoscaling "k8s.io/api/autoscaling/v1"

	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/metric"
	corev1 "k8s.io/api/core/v1"
)

// CustomPodAutoscalerConfig defines the configuration options that can be passed to the CustomPodAutoscaler
type CustomPodAutoscalerConfig struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CustomPodAutoscalerSpec defines the desired state of CustomPodAutoscaler
type CustomPodAutoscalerSpec struct {
	// The image of the Custom Pod Autoscaler
	Template PodTemplateSpec `json:"template"`
	// ScaleTargetRef defining what the Custom Pod Autoscaler should manage
	ScaleTargetRef autoscaling.CrossVersionObjectReference `json:"scaleTargetRef"`
	// Configuration options to be delivered as environment variables to the container
	Config                    []CustomPodAutoscalerConfig `json:"config,omitempty"`
	ProvisionRole             *bool                       `json:"provisionRole,omitempty"`
	ProvisionRoleBinding      *bool                       `json:"provisionRoleBinding,omitempty"`
	ProvisionServiceAccount   *bool                       `json:"provisionServiceAccount,omitempty"`
	ProvisionPod              *bool                       `json:"provisionPod,omitempty"`
	RoleRequiresMetricsServer *bool                       `json:"roleRequiresMetricsServer,omitempty"`
	RoleRequiresArgoRollouts  *bool                       `json:"roleRequiresArgoRollouts,omitempty"`
}

// CustomPodAutoscalerStatus defines the observed state of CustomPodAutoscaler
type CustomPodAutoscalerStatus struct {
	// LastScaleTime is the last time the CustomPodAutoscaler scaled the number of pods, used by the autoscaler to
	// keep track of downscale stabilization between restarts.
	// +optional
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`

	// CurrentReplicas is the current number of replicas of pods managed by this autoscaler, as last seen by the autoscaler.
	// +optional
	CurrentReplicas int32 `json:"currentReplicas"`

	// DesiredReplicas is the desired number of replicas of pods managed by this autoscaler, as last calculated by the
	// autoscaler.
	// +optional
	DesiredReplicas int32 `json:"desiredReplicas"`

	// CurrentMetrics is the last set of metrics gathered by this autoscaler, the last metrics gathered in the
	// metric gathering stage.
	// +optional
	CurrentMetrics []metric.ResourceMetric `json:"currentMetrics"`

	// CurrentEvaluation is the last evaluation decision retrieved by this autoscaler, the last evaluation decision
	// retrieved in the evaluation stage.
	// +optional
	CurrentEvaluation evaluate.Evaluation `json:"currentEvaluation"`

	// Reference is the identifier for the resource being scaled in the format <api-version>/<api-kind/<name>.
	// +optional
	Reference string `json:"reference"`
}

// CustomPodAutoscaler is the Schema for the custompodautoscalers API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cpa
// +kubebuilder:printcolumn:name="Reference",type="string",JSONPath=`.status.reference`,description="The identifier for the resource being scaled in the format <api-version>/<api-kind/<name>"
// +kubebuilder:printcolumn:name="Current Replicas",type="string",JSONPath=`.status.currentReplicas`,description="The current number of replicas of pods managed by this autoscaler as last seen by the autoscaler"
// +kubebuilder:printcolumn:name="Desired Replicas",type="string",JSONPath=`.status.desiredReplicas`,description="The desired number of replicas of pods managed by this autoscaler as last calculated by the autoscaler"
// +kubebuilder:printcolumn:name="Last Scale Time",type="date",JSONPath=`.status.lastScaleTime`,description="The last time the CustomPodAutoscaler scaled the number of pods"
// +groupName=custompodautoscaler.com
type CustomPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPodAutoscalerSpec   `json:"spec,omitempty"`
	Status CustomPodAutoscalerStatus `json:"status,omitempty"`
}

// CustomPodAutoscalerList contains a list of CustomPodAutoscaler
// +kubebuilder:object:root=true
type CustomPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomPodAutoscaler `json:"items"`
}

type PodTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta PodMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// +kubebuilder:pruning:PreserveUnknownFields
type PodMeta metav1.ObjectMeta

type PodSpec corev1.PodSpec

func init() {
	SchemeBuilder.Register(&CustomPodAutoscaler{}, &CustomPodAutoscalerList{})
}
