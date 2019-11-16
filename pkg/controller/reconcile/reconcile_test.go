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

package reconcile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	custompodautoscalerv1alpha1 "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/apis/custompodautoscaler/v1alpha1"
	k8sreconcile "github.com/jthomperoo/custom-pod-autoscaler-operator/pkg/controller/reconcile"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller_custompodautoscaler")

type fakeControllerReferencer struct {
	setControllerReference func(owner, object v1.Object, scheme *runtime.Scheme) error
}

func (f *fakeControllerReferencer) SetControllerReference(owner, object v1.Object, scheme *runtime.Scheme) error {
	return f.setControllerReference(owner, object, scheme)
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
		description string
		expected    reconcile.Result
		expectedErr error
		client      client.Client
		scheme      *runtime.Scheme
		instance    *custompodautoscalerv1alpha1.CustomPodAutoscaler
		obj         metav1.Object
		referencer  func(owner, object v1.Object, scheme *runtime.Scheme) error
	}{
		{
			"Fail to set controller reference",
			reconcile.Result{},
			errors.New("Fail to set controller reference"),
			nil,
			&runtime.Scheme{},
			&custompodautoscalerv1alpha1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			func(owner, object v1.Object, scheme *runtime.Scheme) error {
				return errors.New("Fail to set controller reference")
			},
		},
		{
			"Fail to get object",
			reconcile.Result{},
			errors.New("Fail to get object"),
			func() *fakeClient {
				fclient := &fakeClient{}
				// Client fails to get object
				fclient.get = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					return errors.New("Fail to get object")
				}
				return fclient
			}(), &runtime.Scheme{},
			&custompodautoscalerv1alpha1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			func(owner, object v1.Object, scheme *runtime.Scheme) error {
				return nil
			},
		},
		{
			"Fail to create object",
			reconcile.Result{},
			errors.New("Fail to create object"),
			func() *fakeClient {
				fclient := &fakeClient{}
				// Client reports object not found
				fclient.get = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{}, key.Namespace)
				}
				// Creation fails
				fclient.create = func(ctx context.Context, obj runtime.Object) error {
					return errors.New("Fail to create object")
				}
				return fclient
			}(), &runtime.Scheme{},
			&custompodautoscalerv1alpha1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			func(owner, object v1.Object, scheme *runtime.Scheme) error {
				return nil
			},
		},
		{
			"Successfully create new object",
			reconcile.Result{},
			nil,
			func() *fakeClient {
				fclient := &fakeClient{}
				// Client reports object not found
				fclient.get = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{}, key.Namespace)
				}
				// Creation fails
				fclient.create = func(ctx context.Context, obj runtime.Object) error {
					return nil
				}
				return fclient
			}(), &runtime.Scheme{},
			&custompodautoscalerv1alpha1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			func(owner, object v1.Object, scheme *runtime.Scheme) error {
				return nil
			},
		},
		{
			"Object already exists",
			reconcile.Result{},
			nil,
			fake.NewFakeClientWithScheme(func() *runtime.Scheme {
				s := runtime.NewScheme()
				s.AddKnownTypes(custompodautoscalerv1alpha1.SchemeGroupVersion, &corev1.Pod{
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
			&runtime.Scheme{},
			&custompodautoscalerv1alpha1.CustomPodAutoscaler{},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test pod",
					Namespace: "test namespace",
				},
			},
			func(owner, object v1.Object, scheme *runtime.Scheme) error {
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			reconciler := &k8sreconcile.KubernetesResourceReconciler{
				Client:               test.client,
				Scheme:               test.scheme,
				ControllerReferencer: test.referencer,
			}
			reqLogger := log.WithValues("Request.Namespace", "test", "Request.Name", "test")

			result, err := reconciler.Reconcile(reqLogger, test.instance, test.obj)
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
