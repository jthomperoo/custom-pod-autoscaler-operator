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

package reconcile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/google/go-cmp/cmp"
	custompodautoscalercomv1 "github.com/jthomperoo/custom-pod-autoscaler-operator/api/v1"
	k8sreconcile "github.com/jthomperoo/custom-pod-autoscaler-operator/reconcile"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logr.Discard()

type fakeClient struct {
	get         func(ctx context.Context, key client.ObjectKey, obj client.Object) error
	list        func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	create      func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	delete      func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	update      func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	patch       func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error
	deleteAllOf func(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error
	status      func() client.StatusWriter
	scheme      func() *runtime.Scheme
	restMapper  func() meta.RESTMapper
}

func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return f.get(ctx, key, obj)
}

func (f *fakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return f.list(ctx, list, opts...)
}

func (f *fakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return f.create(ctx, obj, opts...)
}

func (f *fakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return f.delete(ctx, obj, opts...)
}

func (f *fakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return f.update(ctx, obj, opts...)
}

func (f *fakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return f.patch(ctx, obj, patch, opts...)
}

func (f *fakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return f.deleteAllOf(ctx, obj, opts...)
}

func (f *fakeClient) Status() client.StatusWriter {
	return f.status()
}

func (f *fakeClient) Scheme() *runtime.Scheme {
	return f.scheme()
}

func (f *fakeClient) RESTMapper() meta.RESTMapper {
	return f.restMapper()
}

func TestReconcile(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description     string
		expected        reconcile.Result
		expectedErr     error
		reconciler      *k8sreconcile.KubernetesResourceReconciler
		logger          logr.Logger
		instance        *custompodautoscalercomv1.CustomPodAutoscaler
		obj             metav1.Object
		shouldProvision bool
		updatable       bool
	}{
		{
			"Fail to set controller reference",
			reconcile.Result{},
			errors.New("Fail to set controller reference"),
			&k8sreconcile.KubernetesResourceReconciler{
				Client: nil,
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return errors.New("Fail to set controller reference")
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			true,
			false,
		},
		{
			"Fail to get object",
			reconcile.Result{},
			errors.New("Fail to get object"),
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					// Client fails to get object
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New("Fail to get object")
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			true,
			false,
		},
		{
			"Fail to create object",
			reconcile.Result{},
			errors.New("Fail to create object"),
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					// Client reports object not found
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return apierrors.NewNotFound(schema.GroupResource{}, key.Namespace)
					}
					// Creation fails
					fclient.create = func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
						return errors.New("Fail to create object")
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			true,
			false,
		},
		{
			"Success, no object found and don't provision a new one",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					// Client reports object not found
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return apierrors.NewNotFound(schema.GroupResource{}, key.Namespace)
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			false,
			false,
		},
		{
			"Successfully create new object",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					// Client reports object not found
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return apierrors.NewNotFound(schema.GroupResource{}, key.Namespace)
					}
					// Creation successful
					fclient.create = func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
						return nil
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			true,
			false,
		},
		{
			"Object already exists; Pod being deleted, skip updating",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return nil
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
					DeletionTimestamp: &v1.Time{
						Time: time.Now(),
					},
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
			},
			true,
			false,
		},
		{
			"Object already exists; should be provisioned and is updatable, fail to update",
			reconcile.Result{},
			errors.New("Fail to update"),
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return nil
					}
					// Fail to update
					fclient.update = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return errors.New("Fail to update")
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test sa",
					Namespace: "test namespace",
				},
			},
			true,
			true,
		},
		{
			"Object already exists; should be provisioned and is updatable, update success",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return nil
					}
					// Fail to update
					fclient.update = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return nil
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test sa",
					Namespace: "test namespace",
				},
			},
			true,
			true,
		},
		{
			"Object already exists; should be provisioned and isn't updatable, fail to delete",
			reconcile.Result{},
			errors.New("Fail to delete"),
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return nil
					}
					// Fail to delete
					fclient.delete = func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
						return errors.New("Fail to delete")
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test sa",
					Namespace: "test namespace",
				},
			},
			true,
			false,
		},
		{
			"Object already exists; should be provisioned and isn't updatable, delete success",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return nil
					}
					// Fail to delete
					fclient.delete = func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
						return nil
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{},
			&corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test sa",
					Namespace: "test namespace",
				},
			},
			true,
			false,
		},
		{
			"Object already exists with owner not set, fail to update",
			reconcile.Result{},
			errors.New("Fail to update object"),
			&k8sreconcile.KubernetesResourceReconciler{
				Client: func() *fakeClient {
					fclient := &fakeClient{}
					fclient.get = func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return nil
					}
					// Update fails
					fclient.update = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return errors.New("Fail to update object")
					}
					return fclient
				}(),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{
				TypeMeta: v1.TypeMeta{
					Kind:       "custompodautoscaler",
					APIVersion: "apiextensions.k8s.io/v1beta1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "testcpa",
					UID:  "testuid",
				},
			},
			&corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test sa",
					Namespace: "test namespace",
				},
			},
			false,
			false,
		},
		{
			"Object already exists with owner not set, successful update",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: fake.NewFakeClientWithScheme(func() *runtime.Scheme {
					s := runtime.NewScheme()
					s.AddKnownTypes(schema.GroupVersion{
						Group:   "",
						Version: "v1",
					}, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
						},
					})
					return s
				}(),
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
						},
					},
				),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{
				TypeMeta: v1.TypeMeta{
					Kind:       "custompodautoscaler",
					APIVersion: "apiextensions.k8s.io/v1beta1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "testcpa",
					UID:  "testuid",
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			false,
			false,
		},
		{
			"Object already exists with owner set",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: fake.NewFakeClientWithScheme(func() *runtime.Scheme {
					s := runtime.NewScheme()
					s.AddKnownTypes(schema.GroupVersion{
						Group:   "",
						Version: "v1",
					}, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
						},
					})
					return s
				}(),
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test pod",
							Namespace: "test namespace",
						},
					},
				),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{
				TypeMeta: v1.TypeMeta{
					Kind:       "custompodautoscaler",
					APIVersion: "apiextensions.k8s.io/v1beta1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "testcpa",
					UID:  "testuid",
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
					OwnerReferences: []v1.OwnerReference{
						{
							Kind:       "custompodautoscaler",
							APIVersion: "apiextensions.k8s.io/v1beta1",
							Name:       "testcpa",
							UID:        "testuid",
						},
					},
				},
			},
			false,
			false,
		},
		{
			"Service account already exists, retain secret",
			reconcile.Result{},
			nil,
			&k8sreconcile.KubernetesResourceReconciler{
				Client: fake.NewFakeClientWithScheme(func() *runtime.Scheme {
					s := runtime.NewScheme()
					s.AddKnownTypes(schema.GroupVersion{
						Group:   "",
						Version: "v1",
					}, &corev1.ServiceAccount{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ServiceAccount",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test sa",
							Namespace: "test namespace",
						},
						Secrets: []corev1.ObjectReference{
							{
								Name: "test secret",
							},
						},
					})
					return s
				}(),
					&corev1.ServiceAccount{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ServiceAccount",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test sa",
							Namespace: "test namespace",
						},
					},
				),
				Scheme: &runtime.Scheme{},
				ControllerReferencer: func(owner, object v1.Object, scheme *runtime.Scheme) error {
					return nil
				},
			},
			log.WithValues("Request.Namespace", "test", "Request.Name", "test"),
			&custompodautoscalercomv1.CustomPodAutoscaler{
				TypeMeta: v1.TypeMeta{
					Kind:       "custompodautoscaler",
					APIVersion: "apiextensions.k8s.io/v1beta1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "testcpa",
					UID:  "testuid",
				},
			},
			&corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test sa",
					Namespace: "test namespace",
					OwnerReferences: []v1.OwnerReference{
						{
							Kind:       "custompodautoscaler",
							APIVersion: "apiextensions.k8s.io/v1beta1",
							Name:       "testcpa",
							UID:        "testuid",
						},
					},
				},
			},
			true,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.reconciler.Reconcile(test.logger, test.instance, test.obj, test.shouldProvision, test.updatable)
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
