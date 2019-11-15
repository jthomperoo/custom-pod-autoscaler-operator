package reconcile

import (
	"context"

	custompodautoscalerv1alpha1 "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/apis/custompodautoscaler/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type KubernetesResourceReconciler struct {
	Scheme *runtime.Scheme
	Client client.Client
}

// Reconcile manages k8s objects, making sure that the supplied object exists, and if it
// doesn't it creates one
func (k *KubernetesResourceReconciler) Reconcile(
	reqLogger logr.Logger,
	instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
	obj metav1.Object,
) (reconcile.Result, error) {
	// Set CustomPodAutoscaler instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, obj, k.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if k8s object already exists
	runtimeObj := obj.(runtime.Object)
	err := k.Client.Get(context.Background(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, runtimeObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// k8s object doesn't exist, create a new one
			reqLogger.Info("Creating a new k8s object ", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			err = k.Client.Create(context.Background(), runtimeObj)
			if err != nil {
				return reconcile.Result{}, err
			}
			// k8s object created successfully - don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// k8s object already exists - don't requeue
	reqLogger.Info("Skip reconcile: k8s object already exists", "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	return reconcile.Result{}, nil
}
