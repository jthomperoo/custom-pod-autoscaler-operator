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
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	custompodautoscalerv1alpha1 "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/apis/custompodautoscaler/v1alpha1"
	"github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/controller/custompodautoscaler"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	admissiontypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
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
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &custompodautoscalerv1alpha1.CustomPodAutoscaler{})
				return s
			}(),
				&custompodautoscalerv1alpha1.CustomPodAutoscaler{
					Spec: custompodautoscalerv1alpha1.CustomPodAutoscalerSpec{
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
					Spec: custompodautoscalerv1alpha1.CustomPodAutoscalerSpec{
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
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
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
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name: "test container",
									},
								},
							},
						},
						Config: []custompodautoscalerv1alpha1.CustomPodAutoscalerConfig{
							custompodautoscalerv1alpha1.CustomPodAutoscalerConfig{
								Name:  "first env var",
								Value: "first env var value",
							},
							custompodautoscalerv1alpha1.CustomPodAutoscalerConfig{
								Name:  "second env var",
								Value: "second env var value",
							},
							custompodautoscalerv1alpha1.CustomPodAutoscalerConfig{
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
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
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
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
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
					instance *custompodautoscalerv1alpha1.CustomPodAutoscaler,
					obj metav1.Object,
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
			reconciler := &custompodautoscaler.ReconcileCustomPodAutoscaler{
				Client:                       test.client,
				Scheme:                       runtime.NewScheme(),
				KubernetesResourceReconciler: test.k8sreconciler,
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

type fakeManager struct {
	add                 func(manager.Runnable) error
	setFields           func(interface{}) error
	start               func(<-chan struct{}) error
	getConfig           func() *rest.Config
	getScheme           func() *runtime.Scheme
	getAdmissionDecoder func() admissiontypes.Decoder
	getClient           func() client.Client
	getFieldIndexer     func() client.FieldIndexer
	getCache            func() cache.Cache
	getRecorder         func(name string) record.EventRecorder
	getRestMapper       func() meta.RESTMapper
}

// Add sets dependencies on i, and adds it to the list of runnables to start.
func (f *fakeManager) Add(r manager.Runnable) error {
	return f.add(r)
}

func (f *fakeManager) SetFields(i interface{}) error {
	return f.setFields(i)
}

func (f *fakeManager) Start(stop <-chan struct{}) error {
	return f.start(stop)
}

func (f *fakeManager) GetConfig() *rest.Config {
	return f.getConfig()
}

func (f *fakeManager) GetClient() client.Client {
	return f.getClient()
}

func (f *fakeManager) GetScheme() *runtime.Scheme {
	return f.getScheme()
}

func (f *fakeManager) GetAdmissionDecoder() admissiontypes.Decoder {
	return f.getAdmissionDecoder()
}

func (f *fakeManager) GetFieldIndexer() client.FieldIndexer {
	return f.getFieldIndexer()
}

func (f *fakeManager) GetCache() cache.Cache {
	return f.getCache()
}

func (f *fakeManager) GetRecorder(name string) record.EventRecorder {
	return f.getRecorder(name)
}

func (f *fakeManager) GetRESTMapper() meta.RESTMapper {
	return f.getRestMapper()
}

type fakeController struct {
	reconcile func(r reconcile.Request) (reconcile.Result, error)

	watch func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error

	// Start starts the controller.  Start blocks until stop is closed or a controller has an error starting.
	start func(stop <-chan struct{}) error
}

func (f *fakeController) Reconcile(r reconcile.Request) (reconcile.Result, error) {
	return f.reconcile(r)
}

func (f *fakeController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	return f.watch(src, eventhandler, predicates...)
}

func (f *fakeController) Start(stop <-chan struct{}) error {
	return f.start(stop)
}

func TestAdd(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    error
		mgr         manager.Manager
		linker      custompodautoscaler.ControllerLinker
	}{
		{
			"Fail to create new controller",
			errors.New("Fail to create new linker"),
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				return nil, errors.New("Fail to create new linker")
			},
		},
		{
			"Fail to watch for Custom Pod Autoscaler changes",
			errors.New("Fail to watch CPA"),
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				fController := &fakeController{}
				fController.watch = func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
					kind, ok := src.(*source.Kind)
					if ok {
						_, ok := kind.Type.(*custompodautoscalerv1alpha1.CustomPodAutoscaler)
						if ok {
							return errors.New("Fail to watch CPA")
						}
					}
					return nil
				}
				return fController, nil
			},
		},
		{
			"Fail to watch for Pod changes",
			errors.New("Fail to watch pod"),
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				fController := &fakeController{}
				fController.watch = func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
					kind, ok := src.(*source.Kind)
					if ok {
						_, ok := kind.Type.(*corev1.Pod)
						if ok {
							return errors.New("Fail to watch pod")
						}
					}
					return nil
				}
				return fController, nil
			},
		},
		{
			"Fail to watch for Service Account changes",
			errors.New("Fail to watch service account"),
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				fController := &fakeController{}
				fController.watch = func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
					kind, ok := src.(*source.Kind)
					if ok {
						_, ok := kind.Type.(*corev1.ServiceAccount)
						if ok {
							return errors.New("Fail to watch service account")
						}
					}
					return nil
				}
				return fController, nil
			},
		},
		{
			"Fail to watch for Role changes",
			errors.New("Fail to watch role"),
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				fController := &fakeController{}
				fController.watch = func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
					kind, ok := src.(*source.Kind)
					if ok {
						_, ok := kind.Type.(*rbacv1.Role)
						if ok {
							return errors.New("Fail to watch role")
						}
					}
					return nil
				}
				return fController, nil
			},
		},
		{
			"Fail to watch for Role Binding changes",
			errors.New("Fail to watch role binding"),
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				fController := &fakeController{}
				fController.watch = func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
					kind, ok := src.(*source.Kind)
					if ok {
						_, ok := kind.Type.(*rbacv1.RoleBinding)
						if ok {
							return errors.New("Fail to watch role binding")
						}
					}
					return nil
				}
				return fController, nil
			},
		},
		{
			"Successfully add",
			nil,
			func() *fakeManager {
				f := &fakeManager{}
				f.getClient = func() client.Client {
					return fake.NewFakeClient()
				}

				f.getScheme = func() *runtime.Scheme {
					return runtime.NewScheme()
				}
				return f
			}(),
			func(name string, mgr manager.Manager, options controller.Options) (controller.Controller, error) {
				fController := &fakeController{}
				fController.watch = func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
					return nil
				}
				return fController, nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := custompodautoscaler.Add(test.mgr, test.linker)
			if !cmp.Equal(err, test.expected, equateErrorMessage) {
				t.Errorf("Error mismatch (-want +got):\n%s", cmp.Diff(test.expected, err, equateErrorMessage))
				return
			}
		})
	}
}
