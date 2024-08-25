package v1beta1

import "fmt"

func (e *Environment) GetKey() string {
	return fmt.Sprintf("%s/%s-%s.tar.gz", *e.Spec.Source.S3.Prefix, e.GetName(), e.GetNamespace())
}

func (e *Environment) GetBucket() string {
	return *e.Spec.Source.S3.Bucket
}

func (e *Environment) GetRevision() string {
	return e.Spec.Revision
}

func (e *Environment) HasFailed() bool {
	return e.Status.Stage == EnvironmentBuildJobFailed || e.Status.Stage == EnvironmentDeploymentFailed
}

func (e *Environment) IsDeployed() bool {
	return e.Status.DeployedRevision == e.Spec.Revision
}

func (e *Environment) HasDeviated() bool {
	return e.Status.ExpectedRevision != "" && e.Status.ExpectedRevision != e.Spec.Revision
}
