// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	"fmt"
	"net/url"
)

// UseSSL parses the endpoint URI and returns true if the endpoint is
// using SSL.
func (e *EnvironmentS3Spec) UseSSL() bool {
	url, err := url.Parse(*e.Endpoint)
	if err != nil {
		return false
	}
	return url.Scheme == "https"
}

// GetEndpoint returns the endpoint for the S3 service.  Because seaway can
// run against local clusters with a port-forwarded service (either through
// kubectl or ingresses), if the LocalPort field is set, it will return a
// localhost endpoint with the port number.  Otherwise, it will return the
// endpoint as-is.
// TODO: Setting the LocalPort and Endpoint should be mutually exclusive
// as to avoid confusion in a configuration.  Add this validation to the API
// when we have a chance.
func (e *EnvironmentS3Spec) GetEndpoint() string {
	if e.LocalPort != nil {
		return fmt.Sprintf("localhost:%d", *e.LocalPort)
	} else {
		return *e.Endpoint
	}
}
