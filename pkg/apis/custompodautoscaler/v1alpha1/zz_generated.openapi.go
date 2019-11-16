// +build !ignore_autogenerated

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
// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscaler":       schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscaler(ref),
		"./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerConfig": schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscalerConfig(ref),
		"./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerSpec":   schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscalerSpec(ref),
		"./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerStatus": schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscalerStatus(ref),
	}
}

func schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscaler(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CustomPodAutoscaler is the Schema for the custompodautoscalers API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerSpec", "./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscalerConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CustomPodAutoscalerConfig defines the configuration options that can be passed to the CustomPodAutoscaler",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"name": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"value": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"name", "value"},
			},
		},
	}
}

func schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscalerSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CustomPodAutoscalerSpec defines the desired state of CustomPodAutoscaler",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"template": {
						SchemaProps: spec.SchemaProps{
							Description: "The image of the Custom Pod Autoscaler",
							Ref:         ref("k8s.io/api/core/v1.PodTemplateSpec"),
						},
					},
					"scaleTargetRef": {
						SchemaProps: spec.SchemaProps{
							Description: "ScaleTargetRef defining what the Custom Pod Autoscaler should manage",
							Ref:         ref("k8s.io/api/autoscaling/v1.CrossVersionObjectReference"),
						},
					},
					"config": {
						SchemaProps: spec.SchemaProps{
							Description: "Configuration options to be delivered as environment variables to the container",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerConfig"),
									},
								},
							},
						},
					},
				},
				Required: []string{"template", "scaleTargetRef"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/custompodautoscaler/v1alpha1.CustomPodAutoscalerConfig", "k8s.io/api/autoscaling/v1.CrossVersionObjectReference", "k8s.io/api/core/v1.PodTemplateSpec"},
	}
}

func schema_pkg_apis_custompodautoscaler_v1alpha1_CustomPodAutoscalerStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CustomPodAutoscalerStatus defines the observed state of CustomPodAutoscaler",
				Type:        []string{"object"},
			},
		},
	}
}
