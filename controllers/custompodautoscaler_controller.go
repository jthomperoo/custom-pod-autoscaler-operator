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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	custompodautoscalercomv1 "github.com/jthomperoo/custom-pod-autoscaler-operator/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type k8sReconciler interface {
	Reconcile(
		reqLogger logr.Logger,
		instance *custompodautoscalercomv1.CustomPodAutoscaler,
		obj metav1.Object,
		shouldProvision bool,
		updateable bool,
	) (reconcile.Result, error)
}

// CustomPodAutoscalerReconciler reconciles a CustomPodAutoscaler object.
type CustomPodAutoscalerReconciler struct {
	client.Client
	Log                          logr.Logger
	Scheme                       *runtime.Scheme
	KubernetesResourceReconciler k8sReconciler
}

// PrimaryPred is the predicate that filters events for the CustomPodAutoscaler primary resource.
var PrimaryPred = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return true
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
	CreateFunc: func(e event.CreateEvent) bool {
		return true
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return false
	},
}

// SecondaryPred is the predicate that filters events for the CustomPodAutoscaler's secondary
// resources (deployment/service/role/rolebinding).
var SecondaryPred = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return false
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
	CreateFunc: func(e event.CreateEvent) bool {
		return false
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return false
	},
}

// Reconcile reads that state of the cluster for a CustomPodAutoscaler object and makes changes based on the state read
// and what is in the CustomPodAutoscaler.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *CustomPodAutoscalerReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request", req.NamespacedName)

	// Fetch the CustomPodAutoscaler instance
	instance := &custompodautoscalercomv1.CustomPodAutoscaler{}
	err := r.Client.Get(context, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Spec.ProvisionRole == nil {
		defaultVal := true
		instance.Spec.ProvisionRole = &defaultVal
	}
	if instance.Spec.ProvisionRoleBinding == nil {
		defaultVal := true
		instance.Spec.ProvisionRoleBinding = &defaultVal
	}
	if instance.Spec.ProvisionServiceAccount == nil {
		defaultVal := true
		instance.Spec.ProvisionServiceAccount = &defaultVal
	}
	if instance.Spec.ProvisionPod == nil {
		defaultVal := true
		instance.Spec.ProvisionPod = &defaultVal
	}

	// Parse scaleTargetRef
	scaleTargetRef, err := json.Marshal(instance.Spec.ScaleTargetRef)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "custom-pod-autoscaler-operator",
	}

	// Define a new Service Account object
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
	}

	result, err := r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, serviceAccount, *instance.Spec.ProvisionServiceAccount, true)
	if err != nil {
		return result, err
	}

	// Define a new Role object
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "replicationcontrollers", "replicationcontrollers/scale"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "deployments/scale", "replicasets", "replicasets/scale", "statefulsets", "statefulsets/scale"},
				Verbs:     []string{"*"},
			},
		},
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, role, *instance.Spec.ProvisionRole, true)
	if err != nil {
		return result, err
	}

	// Define a new Role Binding object
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     instance.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, roleBinding, *instance.Spec.ProvisionRoleBinding, true)
	if err != nil {
		return result, err
	}

	// Set up Pod labels, if labels are provided in the template Pod Spec the labels are merged
	// with the CPA managed-by label, otherwise only the managed-by label is added
	var podLabels map[string]string
	if instance.Spec.Template.ObjectMeta.Labels == nil {
		podLabels = map[string]string{}
	} else {
		podLabels = instance.Spec.Template.ObjectMeta.Labels
	}
	podLabels["app.kubernetes.io/managed-by"] = "custom-pod-autoscaler-operator"

	// Set up ObjectMeta, if no name or namespaces are provided in the template PodSpec then
	// the CPA name and namespace are used
	objectMeta := instance.Spec.Template.ObjectMeta
	if objectMeta.Name == "" {
		objectMeta.Name = instance.Name
	}
	if objectMeta.Namespace == "" {
		objectMeta.Namespace = instance.Namespace
	}
	objectMeta.Labels = podLabels

	// Set up the PodSpec template
	podSpec := instance.Spec.Template.Spec
	// Inject environment variables to every Container specified by the PodSpec
	containers := []corev1.Container{}
	for _, container := range podSpec.Containers {
		// If no environment variables specified by the template PodSpec, set up empty env vars
		// slice
		var envVars []corev1.EnvVar
		if container.Env == nil {
			envVars = []corev1.EnvVar{}
		} else {
			envVars = container.Env
		}
		// Inject in configuration, such as namespace, target ref and configuration
		// options as environment variables
		envVars = append(envVars, cpaEnvVars(instance, string(scaleTargetRef))...)
		container.Env = envVars
		containers = append(containers, container)
	}
	// Update PodSpec to use the modified containers, and to point to the provisioned service account
	podSpec.Containers = containers
	podSpec.ServiceAccountName = serviceAccount.Name

	// Define Pod object with ObjectMeta and modified PodSpec
	pod := &corev1.Pod{
		ObjectMeta: objectMeta,
		Spec:       podSpec,
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, pod, *instance.Spec.ProvisionPod, false)
	if err != nil {
		return result, err
	}

	return result, nil
}

// cpaEnvVars builds a list of environment variables from the Spec
func cpaEnvVars(cr *custompodautoscalercomv1.CustomPodAutoscaler, scaleTargetRef string) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "scaleTargetRef",
			Value: scaleTargetRef,
		},
		{
			Name:  "namespace",
			Value: cr.Namespace,
		},
	}
	envVars = append(envVars, createEnvVarsFromConfig(cr.Spec.Config)...)
	return envVars
}

// createEnvVarsFromConfig converts CPA config to environment variables
func createEnvVarsFromConfig(configs []custompodautoscalercomv1.CustomPodAutoscalerConfig) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}
	for _, config := range configs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  config.Name,
			Value: config.Value,
		})
	}
	return envVars
}

// SetupWithManager sets up the CustomPodAutoscaler controller, setting up watches with the
// manager provided
func (r *CustomPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&custompodautoscalercomv1.CustomPodAutoscaler{}).
		WithEventFilter(PrimaryPred).
		Owns(&corev1.Pod{}, builder.WithPredicates(SecondaryPred)).
		Owns(&corev1.ServiceAccount{}, builder.WithPredicates(SecondaryPred)).
		Owns(&rbacv1.Role{}, builder.WithPredicates(SecondaryPred)).
		Owns(&rbacv1.RoleBinding{}, builder.WithPredicates(SecondaryPred)).
		Complete(r)
}
