package v1beta1

import (
	"fmt"
	"net/url"
)

func (e *EnvironmentS3Spec) UseSSL() bool {
	url, err := url.Parse(*e.Endpoint)
	if err != nil {
		return false
	}
	return url.Scheme == "https"
}

func (e *EnvironmentS3Spec) GetEndpoint() string {
	if e.LocalPort != nil {
		return fmt.Sprintf("localhost:%d", *e.LocalPort)
	} else {
		return *e.Endpoint
	}
}
