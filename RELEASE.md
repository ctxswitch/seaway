## 0.1.0-pre.10 (2024-08-29)

Continued work towards the initial release.  Expect instability and additional changes.

### Added
* Basic support for service and ingress management based on the presence of port definitions.
* More configuration overrides for build jobs.

### Changed
* The default internal registry now uses the installed minio S3 service as a backend, negating the need to manage PVs on a per cluster basis.
* Build jobs have been consolidated into a single job.
* Install manifests for shared resources reorganized.

### Fixed
* Addresses the issue where jobs would not update status when stages returned an error.