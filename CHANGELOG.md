# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
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

[Unreleased]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.2.1...HEAD
[0.2.1]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/tag/0.1.0