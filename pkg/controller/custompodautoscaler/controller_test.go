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

package custompodautoscaler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	custompodautoscalerv1alpha1 "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/apis/custompodautoscaler/v1alpha1"
	"github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/controller/custompodautoscaler"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type k8sReconciler interface {
	Reconcile(
		reqLogger logr.Logger,
		instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
		obj metav1.Object,
	) (reconcile.Result, error)
}

type fakek8sReconciler struct {
	reconcile func(
		reqLogger logr.Logger,
		instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
		obj metav1.Object,
	) (reconcile.Result, error)
}

func (f *fakek8sReconciler) Reconcile(
	reqLogger logr.Logger,
	instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
	obj metav1.Object,
) (reconcile.Result, error) {
	return f.reconcile(reqLogger, instance, obj)
}

type fakeClient struct {
	get    func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	list   func(ctx context.Context, opts *client.ListOptions, list runtime.Object) error
	create func(ctx context.Context, obj runtime.Object) error
	delete func(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error
	update func(ctx context.Context, obj runtime.Object) error
	status func() client.StatusWriter
}

func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return f.get(ctx, key, obj)
}

func (f *fakeClient) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	return f.list(ctx, opts, list)
}

func (f *fakeClient) Create(ctx context.Context, obj runtime.Object) error {
	return f.create(ctx, obj)
}

func (f *fakeClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	return f.delete(ctx, obj, opts...)
}

func (f *fakeClient) Update(ctx context.Context, obj runtime.Object) error {
	return f.update(ctx, obj)
}

func (f *fakeClient) Status() client.StatusWriter {
	return f.status()
}

func TestReconcile(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description   string
		expected      reconcile.Result
		expectedErr   error
		client        client.Client
		request       reconcile.Request
		k8sreconciler k8sReconciler
	}{
		{
			"No matching CPA",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}()),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			nil,
		},
		{
			"Error on getting CPA",
			reconcile.Result{},
			errors.New("Error getting CPA"),
			func() *fakeClient {
				fclient := &fakeClient{}
				fclient.get = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errors.New("Error getting CPA")
				}
				return fclient
			}(),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			nil,
		},
		{
			"Fail to reconcile service account",
			reconcile.Result{},
			errors.New("Error reconciling service account"),
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				},
			),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			func() *fakek8sReconciler {
				reconciler := &fakek8sReconciler{}
				reconciler.reconcile = func(
					reqLogger logr.Logger,
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
				) (reconcile.Result, error) {
					_, ok := obj.(*corev1.ServiceAccount)
					if ok {
						return reconcile.Result{}, errors.New("Error reconciling service account")
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Fail to reconcile role",
			reconcile.Result{},
			errors.New("Error reconciling role"),
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				},
			),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			func() *fakek8sReconciler {
				reconciler := &fakek8sReconciler{}
				reconciler.reconcile = func(
					reqLogger logr.Logger,
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
				) (reconcile.Result, error) {
					_, ok := obj.(*rbacv1.Role)
					if ok {
						return reconcile.Result{}, errors.New("Error reconciling role")
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Fail to reconcile role binding",
			reconcile.Result{},
			errors.New("Error reconciling rolebinding"),
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				},
			),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			func() *fakek8sReconciler {
				reconciler := &fakek8sReconciler{}
				reconciler.reconcile = func(
					reqLogger logr.Logger,
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
				) (reconcile.Result, error) {
					_, ok := obj.(*rbacv1.RoleBinding)
					if ok {
						return reconcile.Result{}, errors.New("Error reconciling rolebinding")
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Fail to reconcile pod",
			reconcile.Result{},
			errors.New("Error reconciling pod"),
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				},
			),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			func() *fakek8sReconciler {
				reconciler := &fakek8sReconciler{}
				reconciler.reconcile = func(
					reqLogger logr.Logger,
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
				) (reconcile.Result, error) {
					_, ok := obj.(*corev1.Pod)
					if ok {
						return reconcile.Result{}, errors.New("Error reconciling pod")
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Successfully reconcile with default pull policy and no env vars",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				},
			),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			func() *fakek8sReconciler {
				reconciler := &fakek8sReconciler{}
				reconciler.reconcile = func(
					reqLogger logr.Logger,
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
				) (reconcile.Result, error) {
					pod, ok := obj.(*corev1.Pod)
					if ok {
						// Check default pull policy has been set
						if !cmp.Equal(pod.Spec.Containers[0].ImagePullPolicy, corev1.PullIfNotPresent) {
							return reconcile.Result{}, fmt.Errorf(
								"Pull policy mismatch (-want +got):\n%s",
								cmp.Diff(pod.Spec.Containers[0].ImagePullPolicy,
									corev1.PullIfNotPresent),
							)
						}
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Successfully reconcile with no env vars",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
					Spec: custompodautoscalerv1alpha1.CustomPodAutoscalerSpec{
						PullPolicy: corev1.PullPolicy("test pull policy"),
					},
				},
			),
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: "test-namespace",
				},
			},
			func() *fakek8sReconciler {
				reconciler := &fakek8sReconciler{}
				reconciler.reconcile = func(
					reqLogger logr.Logger,
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
				) (reconcile.Result, error) {
					pod, ok := obj.(*corev1.Pod)
					if ok {
						expectedPullPolicy := corev1.PullPolicy("test pull policy")
						// Check test pull policy has been set
						if !cmp.Equal(pod.Spec.Containers[0].ImagePullPolicy, expectedPullPolicy) {
							return reconcile.Result{}, fmt.Errorf(
								"Pull policy mismatch (-want +got):\n%s",
								cmp.Diff(pod.Spec.Containers[0].ImagePullPolicy,
									expectedPullPolicy),
							)
						}
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			reconciler := &custompodautoscaler.ReconcileCustomPodAutoscaler{
				Client:                       test.client,
				Scheme:                       runtime.NewScheme(),
				KubernetesResourceReconciler: test.k8sreconciler,
			}
			result, err := reconciler.Reconcile(test.request)
			if !cmp.Equal(err, test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}

			if !cmp.Equal(result, test.expected) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(result, test.expected))
			}
		})
	}
}
