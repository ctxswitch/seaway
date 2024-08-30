# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.1.0-pre.10 (2024-08-29)

### Added
* Basic support for service and ingress management based on the presence of port definitions.
* More configuration overrides for build jobs.

### Changed
* The default internal registry now uses the installed minio S3 service as a backend, negating the need to manage PVs on a per cluster basis.
* Build jobs have been consolidated into a single job.
* Install manifests for shared resources reorganized.

### Fixed
* Addresses the issue where jobs would not update status when stages returned an error.
