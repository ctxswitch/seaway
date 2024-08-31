### Added
* The seactl `env sync` command now uses the credentials generated in `init shared` to create the secret in the environment namespace for the build job.  This is just a temporary solution to a more wider problem for addressing multitenant environments, but allows for easy install and testing the workflows for single tenant installs.
