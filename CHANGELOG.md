# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.1.0-pre.14 (2024-09-03)

### Added
* Stream container logs through `logs app` and `logs build`
* Services and ingresses are cleaned up when disabled.
* Environment removal through `env clean`.
* The kubernetes context is supported across all commands.
### Changed
* Networking manifest attribute changed to `network`
* The `source/s3` attribute was consolidated to `store`

### Removed
* Setting credentials through the manifest.

### Fixed
* Fixed reconcile loop that was introduced in pre.13
* Tests surfaced several environment stage issues.

## 0.1.0-pre.13 (2024-09-02)

### Fixed
* Environment controller stability fixes.

## 0.1.0-pre.12 (2024-08-31)

### Added
* The seactl `env sync` command now uses the credentials generated in `init shared` to create the secret in the environment namespace for the build job.  This is just a temporary solution to a more wider problem for addressing multitenant environments, but allows for easy install and testing the workflows for single tenant installs.

## 0.1.0-pre.11 (2024-08-30)

### Added
* Resource initialization from the seactl cli.  Very rudimentary right now, but it works.

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
