/*
Copyright 2021 The Custom Pod Autoscaler Authors.

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
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	k8sscale "k8s.io/client-go/scale"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"

	custompodautoscalercomv1 "github.com/jthomperoo/custom-pod-autoscaler-operator/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	managedByLabel           = "app.kubernetes.io/managed-by"
	OwnedByLabel             = "v1.custompodautoscaler.com/owned-by"
	PausedReplicasAnnotation = "v1.custompodautoscaler.com/paused-replicas"
)

type K8sReconciler interface {
	Reconcile(
		reqLogger logr.Logger,
		instance *custompodautoscalercomv1.CustomPodAutoscaler,
		obj metav1.Object,
		shouldProvision bool,
		updateable bool,
	) (reconcile.Result, error)
	PodCleanup(reqLogger logr.Logger, instance *custompodautoscalercomv1.CustomPodAutoscaler) error
}

// CustomPodAutoscalerReconciler reconciles a CustomPodAutoscaler object.
type CustomPodAutoscalerReconciler struct {
	client.Client
	Log                          logr.Logger
	Scheme                       *runtime.Scheme
	KubernetesResourceReconciler K8sReconciler
	ScalingClient                k8sscale.ScalesGetter
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

	// Check the presence of "v1.custompodautoscaler.com/paused-replicas" annotation on the CPA pod
	// Mimics functionality of https://keda.sh/docs/2.11/concepts/scaling-deployments/#pause-autoscaling
	pausedReplicasCount, pausedAnnotationFound := instance.GetAnnotations()[PausedReplicasAnnotation]
	if pausedAnnotationFound {

		// Get paused replicas count from annotation metadata
		pausedReplicasCountInt64, err := strconv.ParseInt(pausedReplicasCount, 10, 32)
		pausedReplicasCountInt32 := int32(pausedReplicasCountInt64)
		if err != nil {
			return reconcile.Result{}, err
		}

		// ScaleTargetRef{} = CrossVersionObjectReference{Kind string, Name string, APIVersion string}s
		// https://github.com/kubernetes/api/blob/v0.27.4/autoscaling/v1/types.go
		scaleTargetRef := instance.Spec.ScaleTargetRef

		// GroupVersion{Group string, Version string}
		// E.X. ParseGroupVersion("custompodautoscaler.com/v1") = GroupVersion{"custompodautoscaler.com", "v1"}
		// https://github.com/kubernetes/apimachinery/blob/v0.27.3/pkg/runtime/schema/group_version.go
		resourceGV, err := schema.ParseGroupVersion(scaleTargetRef.APIVersion)
		if err != nil {
			return reconcile.Result{}, err
		}

		targetGR := schema.GroupResource{
			Group:    resourceGV.Group,    // "custompodautoscaler.com"
			Resource: scaleTargetRef.Kind, // "CustomPodAutoscaler"
		}

		// set up the scaling client if it does not already exist

		////////////////////////////////
		// Consider moving this setup into main
		// I am doing lazy evaluation but might as well put it in main?
		////////////////////////////////

		if r.ScalingClient == nil {
			err = r.SetupScalingClient()
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		// https://github.com/kubernetes/client-go/blob/master/scale/client.go
		// Get the scale request for a resource (https://github.com/kubernetes/api/blob/v0.27.4/autoscaling/v1/types.go)

		////////////////////////////////////////////
		// not sure if it's okay to pass in the regular context instead of context.Background()
		////////////////////////////////////////////

		scaleResource, err := r.ScalingClient.Scales(instance.Namespace).Get(context, targetGR, scaleTargetRef.Name, metav1.GetOptions{})
		if err != nil {
			return reconcile.Result{}, err
		}

		// Set new target replicas
		scaleResource.Spec.Replicas = pausedReplicasCountInt32

		// https://github.com/kubernetes/client-go/blob/master/scale/client.go
		// Update the resource with new replicas
		_, err = r.ScalingClient.Scales(instance.Namespace).Update(context, targetGR, scaleResource, metav1.UpdateOptions{})
		if err != nil {
			return reconcile.Result{}, err
		}

		// Use the reconciler client to delete the pod that normally does the scaling
		if err := r.Client.Delete(context, instance); err != nil {
			return reconcile.Result{}, err
		}

		// Return and don't requeue
		return reconcile.Result{}, nil
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
	if instance.Spec.RoleRequiresMetricsServer == nil {
		defaultVal := false
		instance.Spec.RoleRequiresMetricsServer = &defaultVal
	}
	if instance.Spec.RoleRequiresArgoRollouts == nil {
		defaultVal := false
		instance.Spec.RoleRequiresArgoRollouts = &defaultVal
	}

	// Parse scaleTargetRef
	scaleTargetRef, err := json.Marshal(instance.Spec.ScaleTargetRef)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	labels := map[string]string{
		managedByLabel: "custom-pod-autoscaler-operator",
		OwnedByLabel:   instance.Name,
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

	if *instance.Spec.RoleRequiresMetricsServer {
		role.Rules = append(role.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"metrics.k8s.io", "custom.metrics.k8s.io", "external.metrics.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		})
	}

	if *instance.Spec.RoleRequiresArgoRollouts {
		role.Rules = append(role.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"rollouts", "rollouts/scale"},
			Verbs:     []string{"*"},
		})
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
	podLabels[managedByLabel] = "custom-pod-autoscaler-operator"
	podLabels[OwnedByLabel] = instance.Name

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
		ObjectMeta: metav1.ObjectMeta(objectMeta),
		Spec:       corev1.PodSpec(podSpec),
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, pod, *instance.Spec.ProvisionPod, false)
	if err != nil {
		return result, err
	}

	// Clean up any orphaned pods (e.g. renaming pod, old pod should be deleted)
	err = r.KubernetesResourceReconciler.PodCleanup(reqLogger, instance)
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

// SetupScalingClient sets up a client for the CPA reconciler to use for manually
// setting the replicas count of a scale target pod while the autoscaler is paused.
// Functionality is a simplified version of the autoscaler itself.
func (r *CustomPodAutoscalerReconciler) SetupScalingClient() error {

	// https://github.com/kubernetes/client-go/blob/master/rest/config.go
	// InClusterConfig returns a config object which uses the service account
	// kubernetes gives to pods. It's intended for clients that expect to be
	// running inside a pod running on kubernetes. It will return ErrNotInCluster
	// if called from a process not running in a kubernetes environment.
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	// https://github.com/kubernetes/client-go/blob/master/scale/client.go
	// NewForConfig creates a new ScalesGetter which resolves kinds
	// to resources using the given RESTMapper, and API paths using
	// the given dynamic.APIPathResolverFunc.
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	// https://github.com/kubernetes/client-go/blob/master/restmapper/discovery.go
	// GetAPIGroupResources uses the provided discovery client to gather
	// discovery information and populate a slice of APIGroupResources
	// APIGroupResources{Group metav1.APIGroup, VersionedResources map[string][]metav1.APIResource}
	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return err
	}

	// https://github.com/kubernetes/client-go/blob/master/scale/client.go
	// set up a client for scaling
	scaleClient := k8sscale.New(
		clientset.RESTClient(),
		restmapper.NewDiscoveryRESTMapper(groupResources),
		dynamic.LegacyAPIPathResolverFunc,
		k8sscale.NewDiscoveryScaleKindResolver(
			clientset.Discovery(),
		),
	)

	r.ScalingClient = scaleClient
	return err
}
