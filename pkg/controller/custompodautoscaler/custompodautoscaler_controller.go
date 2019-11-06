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

package custompodautoscaler

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	custompodautoscalerv1alpha1 "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/apis/custompodautoscaler/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

// Add creates a new CustomPodAutoscaler Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCustomPodAutoscaler{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("custompodautoscaler-controller", mgr, controller.Options{Reconciler: r})
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

	return nil
}

// blank assignment to verify that ReconcileCustomPodAutoscaler implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCustomPodAutoscaler{}

// ReconcileCustomPodAutoscaler reconciles a CustomPodAutoscaler object
type ReconcileCustomPodAutoscaler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
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
	err := r.client.Get(context.Background(), request.NamespacedName, instance)
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
		return reconcile.Result{}, err
	}

	// Define a new Service Account object
	serviceAccount := newServiceAccountForCPA(instance)
	result, err := reconcileKubernetesObject(reqLogger, r, instance, serviceAccount)
	if err != nil {
		return result, err
	}

	// Define a new Role object
	role := newRoleForCPA(instance)
	result, err = reconcileKubernetesObject(reqLogger, r, instance, role)
	if err != nil {
		return result, err
	}

	// Define a new Role Binding object
	roleBinding := newRoleBindingForCPA(instance)
	result, err = reconcileKubernetesObject(reqLogger, r, instance, roleBinding)
	if err != nil {
		return result, err
	}

	// Define a new Pod object
	pod := newPodForCPA(instance, string(scaleTargetRef))
	result, err = reconcileKubernetesObject(reqLogger, r, instance, pod)
	if err != nil {
		return result, err
	}

	return result, nil
}

// reconcileKubernetesObject manages k8s objects, making sure that the supplied object exists, and if it
// doesn't it creates one
func reconcileKubernetesObject(
	reqLogger logr.Logger,
	r *ReconcileCustomPodAutoscaler,
	instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
	obj metav1.Object,
) (reconcile.Result, error) {
	// Set CustomPodAutoscaler instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, obj, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if k8s object already exists
	runtimeObj := obj.(runtime.Object)
	err := r.client.Get(context.Background(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, runtimeObj)
	if err != nil && errors.IsNotFound(err) {
		// k8s object doesn't exist, create a new one
		reqLogger.Info("Creating a new k8s object ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		err = r.client.Create(context.Background(), runtimeObj)
		if err != nil {
			return reconcile.Result{}, err
		}
		// k8s object created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// k8s object already exists - don't requeue
	reqLogger.Info("Skip reconcile: k8s object already exists", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	return reconcile.Result{}, nil
}

// newServiceAccountForCPA returns a role with the same name/namespace as the cr for use by the CPA
func newRoleForCPA(cr *custompodautoscalerv1alpha1.CustomPodAutoscaler) *rbacv1.Role {
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "custom-pod-autoscaler",
	}
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
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
}

// newServiceAccountForCPA returns a role binding with the same name/namespace as the cr for use by the CPA
func newRoleBindingForCPA(cr *custompodautoscalerv1alpha1.CustomPodAutoscaler) *rbacv1.RoleBinding {
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "custom-pod-autoscaler",
	}
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      cr.Name,
				Namespace: cr.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     cr.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

// newServiceAccountForCPA returns a service account with the same name/namespace as the cr for use by the CPA
func newServiceAccountForCPA(cr *custompodautoscalerv1alpha1.CustomPodAutoscaler) *corev1.ServiceAccount {
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "custom-pod-autoscaler",
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
	}
}

// newPodForCPA returns a pod with the same name/namespace as the cr with the image specified for use by the CPA
func newPodForCPA(cr *custompodautoscalerv1alpha1.CustomPodAutoscaler, scaleTargetRef string) *corev1.Pod {
	// default pull policy is PullIfNotPresent
	pullPolicy := corev1.PullIfNotPresent
	if cr.Spec.PullPolicy != "" {
		pullPolicy = cr.Spec.PullPolicy
	}
	labels := map[string]string{
		"app":                          cr.Name,
		"app.kubernetes.io/managed-by": "custom-pod-autoscaler",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: cr.Name,
			Containers: []corev1.Container{
				{
					Name:            cr.Name,
					Image:           cr.Spec.Image,
					ImagePullPolicy: pullPolicy,
					Env:             newEnvVars(cr, scaleTargetRef),
				},
			},
		},
	}
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
