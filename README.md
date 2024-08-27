# Seaway

## Next
* Comments and documentation.
* I think I want to go to a single build job that is deleted before starting if it exists, but that's lower priority.
* Container builds and install instructions via kustomize and remote directories
* Initial tests.
* Move to ctxswitch
* Service creation for apps if ports are defined.
* Job build command/argument overrides.

## TODO
* We should potentially create a temporary namespace to keep other users from clobbering environments.
* Add deps directory config.
* Potentially change `source` to `target` or something else a bit more descriptive.  Source is confusing as source is what you are actually working on.
* On sync, check to see if there are old jobs running and remove.  This may relate to moving back to revisionless jobs.
* Consider removing json hints since we use the json fields as they would default
* Add a workqueue for environment reconciliation (P2)
  * Have workers subscribe to the channel and put watch resources on the channel once they come in.
  * Workers will be responsible for the build and deploy.
  * Workers will grab a semephore lock and process.
* Implement env.clean and deps.delete
* Allow seactl to set the kubernetes context so we can ensure that the correct one is being accessed.  We can also allow `--context` as an option for not permanently setting this.
* Registries can use S3 as persistent storage.  Use that by default so we can simplify the deploy (and we have minio installed by default)
* Redo the dependency kustomize tree to seperate it fully from the dev deps.

## Known Bugs
* If there is an error with the job, we aren't failing correctly.  Doesn't update the status and the gives a successful "Revision has been deployed".  I can simulate a failure by setting the platform to `amd64/darwin`
* I've turned off cobra errors so there's no feedback
* Unable to create archive: read config/base: is a directory

2