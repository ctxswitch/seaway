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

package registry

import (
	"io"
	"net/http"
)

type ClientError struct {
	URL        string
	StatusCode int
	Message    string
}

// NewClientError creates a new ClientError from an http.Response.
func NewClientError(resp *http.Response) *ClientError {
	msg, error := io.ReadAll(resp.Body)
	if error != nil {
		msg = []byte("unable to read response body")
	}
	return &ClientError{
		URL:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Message:    string(msg),
	}
}

// Error returns a string representation of the error message.
func (e ClientError) Error() string {
	return e.Message
}

// IsNotFound returns true if the error is a 404.
func IsNotFound(err error) bool {
	switch e := err.(type) { //nolint:gocritic
	case *ClientError:
		return e.StatusCode == http.StatusNotFound
	}

	return false
}
