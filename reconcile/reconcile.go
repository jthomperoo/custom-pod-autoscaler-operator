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

package reconcile

import (
	"context"

	"github.com/go-logr/logr"
	custompodautoscalercomv1 "github.com/jthomperoo/custom-pod-autoscaler-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controllerReferencer func(owner, object v1.Object, scheme *runtime.Scheme) error

// KubernetesResourceReconciler handles reconciling Kubernetes resources, such as pods, service accounts etc.
type KubernetesResourceReconciler struct {
	Scheme               *runtime.Scheme
	Client               client.Client
	ControllerReferencer controllerReferencer
}

// Reconcile manages k8s objects, making sure that the supplied object exists, and if it
// doesn't it creates one
func (k *KubernetesResourceReconciler) Reconcile(
	reqLogger logr.Logger,
	instance *custompodautoscalercomv1.CustomPodAutoscaler,
	obj metav1.Object,
	shouldProvision bool,
	updatable bool,
) (reconcile.Result, error) {
	runtimeObj := obj.(client.Object)
	// Set CustomPodAutoscaler instance as the owner and controller
	err := k.ControllerReferencer(instance, obj, k.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if k8s object already exists
	existingObject := runtimeObj
	err = k.Client.Get(context.Background(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existingObject)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		// Object does not exist
		if !shouldProvision {
			reqLogger.Info("Object not found, no provisioning of resource ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			// Should not provision a new object, wait for existing
			return reconcile.Result{}, nil
		}
		// Should provision, create a new object
		reqLogger.Info("Creating a new k8s object ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		err = k.Client.Create(context.Background(), runtimeObj)
		if err != nil {
			return reconcile.Result{}, err
		}
		// K8s object created successfully - don't requeue
		return reconcile.Result{}, nil
	}

	if existingObject.GetObjectKind().GroupVersionKind().Group == "" &&
		existingObject.GetObjectKind().GroupVersionKind().Version == "v1" &&
		existingObject.GetObjectKind().GroupVersionKind().Kind == "Pod" {
		pod := existingObject.(*corev1.Pod)
		if !pod.ObjectMeta.DeletionTimestamp.IsZero() {
			reqLogger.Info("Pod currently being deleted ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			return reconcile.Result{}, nil
		}
	}

	// Object already exists, update
	if shouldProvision {
		// Only update if object should be provisioned
		if updatable {
			reqLogger.Info("Updating k8s object ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			if existingObject.GetObjectKind().GroupVersionKind().Group == "" &&
				existingObject.GetObjectKind().GroupVersionKind().Version == "v1" &&
				existingObject.GetObjectKind().GroupVersionKind().Kind == "ServiceAccount" {
				reqLogger.Info("Service Account update, retaining secrets ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
				serviceAccount := existingObject.(*corev1.ServiceAccount)
				updatedServiceAccount := runtimeObj.(*corev1.ServiceAccount)
				updatedServiceAccount.Secrets = serviceAccount.Secrets
			}
			// If object can be updated
			err = k.Client.Update(context.Background(), runtimeObj)
			if err != nil {
				return reconcile.Result{}, err
			}
			// Successful update, don't requeue
			return reconcile.Result{}, nil
		}
		reqLogger.Info("Deleting k8s object ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())

		// If object can't be updated, delete and make new
		err = k.Client.Delete(context.Background(), existingObject)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	// Object should not be provisioned, instead update owner reference of
	// existing object
	obj = existingObject.(metav1.Object)
	// Check if CPA set as K8s object owner
	ownerReferences := obj.GetOwnerReferences()
	cpaOwner := false
	for _, owner := range ownerReferences {
		if owner.Kind == instance.Kind && owner.APIVersion == instance.APIVersion && owner.Name == instance.Name {
			cpaOwner = true
			break
		}
	}

	if !cpaOwner {
		reqLogger.Info("CPA not set as owner, updating owner reference", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		ownerReferences = append(ownerReferences, metav1.OwnerReference{
			APIVersion: instance.APIVersion,
			Kind:       instance.Kind,
			Name:       instance.Name,
			UID:        instance.UID,
		})
		obj.SetOwnerReferences(ownerReferences)
		err = k.Client.Update(context.Background(), existingObject)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Skip reconcile: k8s object already exists with expected owner", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	return reconcile.Result{}, nil
}
