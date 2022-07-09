# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- New status fields to help debugging and to give more context for scaling decisions.
    - `lastScaleTime` is the last time the CustomPodAutoscaler scaled the number of pods, used by the autoscaler to
    keep track of downscale stabilization between restarts.
	- `currentReplicas` is the current number of replicas of pods managed by this autoscaler, as last seen by the
    autoscaler.
	- `desiredReplicas` is the desired number of replicas of pods managed by this autoscaler, as last calculated by the
    autoscaler.
	- `currentMetrics` is the last set of metrics gathered by this autoscaler, the last metrics gathered in the
	metric gathering stage
	- `currentEvaluation` is the last evaluation decision retrieved by this autoscaler, the last evaluation decision
	retrieved in the evaluation stage
	- `reference` is the identifier for the resource being scaled in the format `<api-version>/<api-kind/<name>`.
- Certain fields are now added as printer columns, so are visible when you do `kubectl get cpa` to add more context:
    - `Last Scale Time` is the last time the CustomPodAutoscaler scaled the number of pods.
    - `Current Replicas` is the current number of replicas of pods managed by this autoscaler.
    - `Desired Replicas` is the desired number of replicas of pods managed by this autoscaler.
    - `Reference` is the identifier for the resource being scaled in the format `<api-version>/<api-kind/<name>`.

## [v1.3.0] - 2022-07-08
### Fixed
- Bug that pod metadata was not preserved when creating the pod ([#87](https://github.com/jthomperoo/custom-pod-autoscaler-operator/issues/87)).
### Changed
- Each resource provisioned by the CPA will now have the label `v1.custompodautoscaler.com/owned-by` which will
contain the name of the CPA that owns the resource. This is used to help look up resources and link them back to
the owned CPA. This allows the managed Pod to change name, and the operator will know to delete the old Pod when it
provisions the new Pod. Addresses [#95](https://github.com/jthomperoo/custom-pod-autoscaler-operator/issues/95).

## [v1.2.1] - 2022-04-17
### Fixed
- Fixed issue with namespaced deploys not working due to invalid permissions when watching resources in a namespace.

## [v1.2.0] - 2021-12-27
### Added
- Support for Argo Rollouts, new option `roleRequiresArgoRollouts` which will add the required Role permissions to
manage a `argoproj.io/v1alpha1` `Rollout`.

## [v1.1.1] - 2021-06-20
### Fixed
- Bug that when provisioning a role that requires metrics server access (`roleRequiresMetricsServer`) the operator
would not provision a role that had permission to access custom or external metrics.

## [v1.1.0] - 2021-04-07
### Added
- New `roleRequiresMetricsServer` (defaults to `false`), if set to `true` the provisioned role will include permission
to acess the Kubernetes metrics server.

## [v1.0.3] - 2021-03-17
### Changed
- Upgrade to stable Operator SDK version `v1.5.0`.
### Fixed
- Bug that did not allow the operator to install correctly on K8s >= `v1.18.x`

## [v1.0.2] - 2020-09-13
### Fixed
- Bug where the service account was regenerating its secrets on every reconcile, resulting in a pile up of secrets
that are never garbage collected. Service accounts now retain secrets between reconciles.

## [v1.0.1] - 2020-08-15
### Fixed
- When deploying using Cluster wide scope to a namespace, the `ClusterRoleBinding` no longer only searches in the
`default` namespace for the `ServiceAccount` - instead it searches in the namespace that the helm chart is deployed to.
This change is only for the helm deploys, as such the kubectl deployment method is no longer recommended, as it only
supports deploying to the `default` namespace for cluster wide installs.

## [v1.0.0] - 2020-07-19

## [v0.7.0] - 2020-07-18
### Changed
- CustomPodAutoscaler resource changed from `v1alpha` to `v1`.
- Operator Docker image now using `distroless` rather than RHEL `ubi7`.
- Add deployment via helm.

## [v0.6.0] - 2020-06-24
### Added
- New options for deciding if a resource should be provisioned by the CPAO, or if they are provided already by the
user/another system. All provision options default to `true`.
    - `provisionRole` - determines if a `Role` should be provisioned.
    - `provisionRoleBinding` - determines if a `RoleBinding` should be provisioned.
    - `provisionServiceAccount` - determines if a `ServiceAccount` should beprovisioned.
    - `provisionPod` - determines if a `Pod` should be provisioned.
- Resources can now be updated at runtime, without deleting and recreating the CPA.
    - All resources will be updated using the standard K8s Update procedure, except for `Pod` resources, which will be
    deleted and recreated, in order to use any new image provided.

## [v0.5.0] - 2020-01-18
### Added
- Add permissions to role for managing ReplicationControllers, ReplicaSets, and StatefulSets.
- Add permissions to use scaling API.
- When a resource already exists, the operator will check if the assigned CPA has been set as its owner; if it isn't it
will set it, if not it will skip it. This can be used by CPAs to modify the resources for the CPA.

## [0.4.0] - 2019-11-16
### Changed
- Use ScaleTargetRef rather than a label selector to choose which resource to manage, consistent with Horizontal Pod
Autoscaler.
- Can now define a PodSpec rather than a Docker image.
### Removed
- PullPolicy removed as it can now be defined in the template PodSpec.
- Image removed as it is now defined within a PodSpec instead.

## [0.3.0] - 2019-11-03
### Changed
- Deliver configuration via environment variables in a consistent format with YAML config, all lowercase.

## [0.2.1] - 2019-10-30
### Fixed
- Update operator YAML for cluster install to point to correct service account.

## [0.2.0] - 2019-10-28
### Added
- Allow single line install of operator.
- Allow specifying selector in custom resource defintion file.
- Allow operator to run either cluster-scoped or namespace-scoped.
- Create service account, role and role binding for CPA.
- Support creating a CPA in a namespace.
### Changed
- Added permissions for operator to manage roles and role bindings in operator role definition.

## [0.1.0] - 2019-09-28
### Added
- Allow use of config.yaml to select deployment to manage.
- Provide YAML config for Custom Pod Autoscaler custom resource.
- Provide YAML config for resources required to run operator.
- Provide controller that provisions a single pod deployment to run CPA in.
- Allow creation/deletion of CPA.

[Unreleased]:
https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.3.0...HEAD
[v1.3.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.2.1...v1.3.0
[v1.2.1]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.1.1...v1.2.0
[v1.1.1]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.0.3...v1.1.0
[v1.0.3]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.0.2...v1.0.3
[v1.0.2]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v0.7.0...v1.0.0
[v0.7.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.4.0...v0.5.0
[0.4.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.3.0...0.4.0
[0.3.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.2.1...0.3.0
[0.2.1]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/tag/0.1.0
