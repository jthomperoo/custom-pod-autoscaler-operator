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
// +build unit

package custompodautoscaler

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	custompodautoscalerv1alpha1 "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/apis/custompodautoscaler/v1alpha1"
	k8sreconcile "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/controller/reconcile"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_custompodautoscaler")

type k8sReconciler interface {
	Reconcile(
		reqLogger logr.Logger,
		instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
		obj metav1.Object,
	) (reconcile.Result, error)
}

// ReconcileCustomPodAutoscaler reconciles a CustomPodAutoscaler object
type ReconcileCustomPodAutoscaler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client                       client.Client
	Scheme                       *runtime.Scheme
	KubernetesResourceReconciler k8sReconciler
}

// ControllerLinker is used to create a new controller linked to the manager provided
type ControllerLinker func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error)

// Add creates a new CustomPodAutoscaler Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, linker ControllerLinker) error {
	// Set up reconciler
	client := mgr.GetClient()
	scheme := mgr.GetScheme()
	r := &ReconcileCustomPodAutoscaler{
		Client: client,
		Scheme: scheme,
		KubernetesResourceReconciler: &k8sreconcile.KubernetesResourceReconciler{
			Client:               client,
			Scheme:               scheme,
			ControllerReferencer: controllerutil.SetControllerReference,
		},
	}

	// Create a new controller
	c, err := linker("custompodautoscaler-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CustomPodAutoscaler
	err = c.Watch(&source.Kind{Type: &custompodautoscalerv1alpha1.CustomPodAutoscaler{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner CustomPodAutoscaler
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &custompodautoscalerv1alpha1.CustomPodAutoscaler{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ServiceAccounts and requeue the owner CustomPodAutoscaler
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &custompodautoscalerv1alpha1.CustomPodAutoscaler{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Roles and requeue the owner CustomPodAutoscaler
	err = c.Watch(&source.Kind{Type: &rbacv1.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &custompodautoscalerv1alpha1.CustomPodAutoscaler{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RoleBindings and requeue the owner CustomPodAutoscaler
	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &custompodautoscalerv1alpha1.CustomPodAutoscaler{},
	})
	if err != nil {
		return err
	}

	return nil
}

// Reconcile reads that state of the cluster for a CustomPodAutoscaler object and makes changes based on the state read
// and what is in the CustomPodAutoscaler.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCustomPodAutoscaler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CustomPodAutoscaler")

	// Fetch the CustomPodAutoscaler instance
	instance := &custompodautoscalerv1alpha1.CustomPodAutoscaler{}
	err := r.Client.Get(context.Background(), request.NamespacedName, instance)
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

	// Parse scaleTargetRef
	scaleTargetRef, err := json.Marshal(instance.Spec.ScaleTargetRef)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "custom-pod-autoscaler",
	}

	// Define a new Service Account object
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
	}
	result, err := r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, serviceAccount)
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
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
				Verbs:     []string{"*"},
			},
		},
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, role)
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
			rbacv1.Subject{
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
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, roleBinding)
	if err != nil {
		return result, err
	}

	labels["app"] = instance.Name

	// default pull policy is PullIfNotPresent
	pullPolicy := corev1.PullIfNotPresent
	if instance.Spec.PullPolicy != "" {
		pullPolicy = instance.Spec.PullPolicy
	}

	// Define a new Pod object
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: instance.Name,
			Containers: []corev1.Container{
				{
					Name:            instance.Name,
					Image:           instance.Spec.Image,
					ImagePullPolicy: pullPolicy,
					Env:             newEnvVars(instance, string(scaleTargetRef)),
				},
			},
		},
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, pod)
	if err != nil {
		return result, err
	}

	return result, nil
}

// newEnvVars builds a list of environment variables from the Spec
func newEnvVars(cr *custompodautoscalerv1alpha1.CustomPodAutoscaler, scaleTargetRef string) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "scaleTargetRef",
			Value: scaleTargetRef,
		},
		corev1.EnvVar{
			Name:  "namespace",
			Value: cr.Namespace,
		},
	}
	envVars = append(envVars, createEnvVarsFromConfig(cr.Spec.Config)...)
	return envVars
}

// createEnvVarsFromConfig converts CPA config to environment variables
func createEnvVarsFromConfig(configs []custompodautoscalerv1alpha1.CustomPodAutoscalerConfig) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}
	for _, config := range configs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  config.Name,
			Value: config.Value,
		})
	}
	return envVars
}
