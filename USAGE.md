# Usage

This will show the basic usage of the Custom Pod Autoscaler Operator, for more
indepth examples check out the 
[Custom Pod Autoscaler repo](https://github.com/jthomperoo/custom-pod-autoscaler).

## Simple Custom Pod Autoscaler

```yaml
apiVersion: custompodautoscaler.com/v1alpha1
kind: CustomPodAutoscaler
metadata:
  name: python-custom-autoscaler
spec:
  template:
    spec:
      containers:
      - name: python-custom-autoscaler
        image: python-custom-autoscaler:latest
        imagePullPolicy: Always
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: hello-kubernetes
  config: 
    - name: interval
      value: "10000"
```

This is a simple Custom Pod Autoscaler, using an image called 
`python-custom-autoscaler:latest`. 

It provides the configuration option `interval` with a value of `10000` as an 
environment variable injected into the container.

The target of this CPA is defined by `scaleTargetRef` - it targets a `Deployment` 
called `hello-kubernetes`.

For more indepth examples check out the 
[Custom Pod Autoscaler repo](https://github.com/jthomperoo/custom-pod-autoscaler).

## Using Custom Resources

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: python-custom-autoscaler
  annotations:
    myCustomAnnotation: test
---
apiVersion: custompodautoscaler.com/v1alpha1
kind: CustomPodAutoscaler
metadata:
  name: python-custom-autoscaler
spec:
  template:
    spec:
      containers:
      - name: python-custom-autoscaler
        image: python-custom-autoscaler:latest
        imagePullPolicy: Always
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: hello-kubernetes
  provisionServiceAccount: false
  config: 
    - name: interval
      value: "10000"
```

This is a Custom Pod Autoscaler that is similar to the basic one defined above, except
that it uses a custom `ServiceAccount`, with the annotation `myCustomAnnotation`.

Take note of the option inside the CPA `provisionServiceAccount: false`, which informs
the CPAO that the user will be providing their own `ServiceAccount`, so it should
not override it with its own provisioned `ServiceAccount`.

This custom resource provision is supported for all resources the CPAO manages:

- `provisionRole` - determines if a `Role` should be provisioned.
- `provisionRoleBinding` - determines if a `RoleBinding` should be provisioned.
- `provisionServiceAccount` - determines if a `ServiceAccount` should be 
provisioned.
- `provisionPod` - determines if a `Pod` should be provisioned.