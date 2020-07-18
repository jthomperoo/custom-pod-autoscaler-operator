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

package controllers_test

import (
	"context"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	custompodautoscalercomv1 "github.com/jthomperoo/custom-pod-autoscaler-operator/api/v1"
	"github.com/jthomperoo/custom-pod-autoscaler-operator/controllers"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestPrimaryPredicate(t *testing.T) {
	result := controllers.PrimaryPred.Create(event.CreateEvent{})
	if !cmp.Equal(result, true) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, true))
		return
	}
	result = controllers.PrimaryPred.Delete(event.DeleteEvent{})
	if !cmp.Equal(result, true) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, true))
		return
	}
	result = controllers.PrimaryPred.Update(event.UpdateEvent{})
	if !cmp.Equal(result, true) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, true))
		return
	}
	result = controllers.PrimaryPred.Generic(event.GenericEvent{})
	if !cmp.Equal(result, false) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, false))
		return
	}
}

func TestSecondaryPredicate(t *testing.T) {
	result := controllers.SecondaryPred.Create(event.CreateEvent{})
	if !cmp.Equal(result, false) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, false))
		return
	}
	result = controllers.SecondaryPred.Delete(event.DeleteEvent{})
	if !cmp.Equal(result, true) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, true))
		return
	}
	result = controllers.SecondaryPred.Update(event.UpdateEvent{})
	if !cmp.Equal(result, false) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, false))
		return
	}
	result = controllers.SecondaryPred.Generic(event.GenericEvent{})
	if !cmp.Equal(result, false) {
		t.Errorf("Boolean mismatch (-want +got):\n%s", cmp.Diff(result, false))
		return
	}
}

type k8sReconciler interface {
	Reconcile(
		reqLogger logr.Logger,
		instance *custompodautoscalercomv1.CustomPodAutoscaler,
		obj metav1.Object,
		shouldProvision bool,
		updatable bool,
	) (reconcile.Result, error)
}

type fakek8sReconciler struct {
	reconcile func(
		reqLogger logr.Logger,
		instance *custompodautoscalercomv1.CustomPodAutoscaler,
		obj metav1.Object,
		shouldProvision bool,
		updatable bool,
	) (reconcile.Result, error)
}

func (f *fakek8sReconciler) Reconcile(
	reqLogger logr.Logger,
	instance *custompodautoscalercomv1.CustomPodAutoscaler,
	obj metav1.Object,
	shouldProvision bool,
	updatable bool,
) (reconcile.Result, error) {
	return f.reconcile(reqLogger, instance, obj, shouldProvision, updatable)
}

type fakeClient struct {
	get         func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	list        func(ctx context.Context, list runtime.Object, opts ...client.ListOption) error
	create      func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error
	delete      func(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error
	update      func(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error
	patch       func(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error
	deleteAllOf func(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error
	status      func() client.StatusWriter
}

func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return f.get(ctx, key, obj)
}

func (f *fakeClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return f.list(ctx, list, opts...)
}

func (f *fakeClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	return f.create(ctx, obj, opts...)
}

func (f *fakeClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	return f.delete(ctx, obj, opts...)
}

func (f *fakeClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	return f.update(ctx, obj, opts...)
}

func (f *fakeClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	return f.patch(ctx, obj, patch, opts...)
}

func (f *fakeClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	return f.deleteAllOf(ctx, obj, opts...)
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
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
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
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
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
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
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
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
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
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
					Spec: custompodautoscalercomv1.CustomPodAutoscalerSpec{
						Template: corev1.PodTemplateSpec{},
					},
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
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
			"Successfully reconcile with no env vars",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
					Spec: custompodautoscalercomv1.CustomPodAutoscalerSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name: "test container",
									},
								},
							},
						},
					},
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
				) (reconcile.Result, error) {
					pod, ok := obj.(*corev1.Pod)
					if ok {
						// Default env vars
						expectedEnvVars := []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "scaleTargetRef",
								Value: `{"kind":"","name":""}`,
							},
							corev1.EnvVar{
								Name:  "namespace",
								Value: "test-namespace",
							},
						}

						if !cmp.Equal(expectedEnvVars, pod.Spec.Containers[0].Env) {
							t.Errorf("Env vars mismatch (-want +got):\n%s",
								cmp.Diff(expectedEnvVars, pod.Spec.Containers[0].Env))
							return reconcile.Result{}, nil
						}
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Successfully reconcile with env vars",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
					Spec: custompodautoscalercomv1.CustomPodAutoscalerSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name: "test container",
									},
								},
							},
						},
						Config: []custompodautoscalercomv1.CustomPodAutoscalerConfig{
							custompodautoscalercomv1.CustomPodAutoscalerConfig{
								Name:  "first env var",
								Value: "first env var value",
							},
							custompodautoscalercomv1.CustomPodAutoscalerConfig{
								Name:  "second env var",
								Value: "second env var value",
							},
							custompodautoscalercomv1.CustomPodAutoscalerConfig{
								Name:  "third env var",
								Value: "third env var value",
							},
						},
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
				) (reconcile.Result, error) {
					pod, ok := obj.(*corev1.Pod)
					if ok {
						expectedEnvVars := []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "scaleTargetRef",
								Value: `{"kind":"","name":""}`,
							},
							corev1.EnvVar{
								Name:  "namespace",
								Value: "test-namespace",
							},
							corev1.EnvVar{
								Name:  "first env var",
								Value: "first env var value",
							},
							corev1.EnvVar{
								Name:  "second env var",
								Value: "second env var value",
							},
							corev1.EnvVar{
								Name:  "third env var",
								Value: "third env var value",
							},
						}

						if !cmp.Equal(expectedEnvVars, pod.Spec.Containers[0].Env) {
							t.Errorf("Env vars mismatch (-want +got):\n%s",
								cmp.Diff(expectedEnvVars, pod.Spec.Containers[0].Env))
							return reconcile.Result{}, nil
						}
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Successfully reconcile with labels set in the container",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
					Spec: custompodautoscalercomv1.CustomPodAutoscalerSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name: "test container",
									},
								},
							},
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"test-label": "test",
								},
							},
						},
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
				) (reconcile.Result, error) {
					pod, ok := obj.(*corev1.Pod)
					if ok {
						expectedEnvVars := []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "scaleTargetRef",
								Value: `{"kind":"","name":""}`,
							},
							corev1.EnvVar{
								Name:  "namespace",
								Value: "test-namespace",
							},
						}

						if !cmp.Equal(expectedEnvVars, pod.Spec.Containers[0].Env) {
							t.Errorf("Env vars mismatch (-want +got):\n%s",
								cmp.Diff(expectedEnvVars, pod.Spec.Containers[0].Env))
							return reconcile.Result{}, nil
						}
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, nil
				}
				return reconciler
			}(),
		},
		{
			"Successfully reconcile with env vars set in pod spec and no config env vars",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalercomv1.GroupVersion, &custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
				})
				return s
			}(),
				&custompodautoscalercomv1.CustomPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-namespace",
					},
					Spec: custompodautoscalercomv1.CustomPodAutoscalerSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name: "test container",
										Env: []corev1.EnvVar{
											corev1.EnvVar{
												Name:  "test container env name",
												Value: "test container env value",
											},
										},
									},
								},
							},
						},
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
					instance *custompodautoscalercomv1.CustomPodAutoscaler,
					obj metav1.Object,
					shouldProvision bool,
					updatable bool,
				) (reconcile.Result, error) {
					pod, ok := obj.(*corev1.Pod)
					if ok {
						expectedEnvVars := []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "test container env name",
								Value: "test container env value",
							},
							corev1.EnvVar{
								Name:  "scaleTargetRef",
								Value: `{"kind":"","name":""}`,
							},
							corev1.EnvVar{
								Name:  "namespace",
								Value: "test-namespace",
							},
						}

						if !cmp.Equal(expectedEnvVars, pod.Spec.Containers[0].Env) {
							t.Errorf("Env vars mismatch (-want +got):\n%s",
								cmp.Diff(expectedEnvVars, pod.Spec.Containers[0].Env))
							return reconcile.Result{}, nil
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
			reconciler := &controllers.CustomPodAutoscalerReconciler{
				Client:                       test.client,
				Scheme:                       runtime.NewScheme(),
				KubernetesResourceReconciler: test.k8sreconciler,
				Log:                          logf.Log.WithName("controller_custompodautoscaler"),
			}
			result, err := reconciler.Reconcile(test.request)
			if !cmp.Equal(err, test.expectedErr, equateErrorMessage) {
				t.Errorf("Error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}

			if !cmp.Equal(result, test.expected) {
				t.Errorf("Result mismatch (-want +got):\n%s", cmp.Diff(result, test.expected))
			}
		})
	}
}
