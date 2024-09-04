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