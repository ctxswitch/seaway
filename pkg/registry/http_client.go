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
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const (
	DefaultHTTPTimeout = 3 * time.Second
)

// Connector is an interface for making HTTP requests to a registry.
type Connector interface {
	Post(url string, data any, v any) error
	Get(url string, v any) error
	Patch(url string, data any, v any) error
}

// HTTPClient is a simple HTTP client that implements the Connector interface.
type HTTPClient struct {
	client http.Client
}

// NewHTTPClient creates a new HTTPClient.
func NewHTTPClient() *HTTPClient {
	client := http.Client{
		// TODO: pull these out as constants
		Timeout: DefaultHTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    60 * time.Second,
			DisableKeepAlives:  false,
			DisableCompression: false,
		},
	}
	return &HTTPClient{
		client: client,
	}
}

// Post makes a POST request to the given URL with the given data and decodes the response
// into the interface that is passed in.
func (h *HTTPClient) Post(url string, data any, v any) error {
	return h.Do(http.MethodPost, url, data, v)
}

// Get makes a GET request to the given URL and decodes the response into the interface
// that is passed in.
func (h *HTTPClient) Get(url string, v any) error {
	return h.Do(http.MethodGet, url, nil, v)
}

// Patch makes a PATCH request to the given URL with the given data and decodes the response
// into the interface that is passed in.
func (h *HTTPClient) Patch(url string, data any, v any) error {
	return h.Do(http.MethodPatch, url, data, v)
}

// Do makes an HTTP request to the given URL with the given method and data and decodes the
// response into the interface that is passed in.
func (h *HTTPClient) Do(method string, url string, data any, out any) error {
	// TODO: clean me up
	var body []byte
	var err error
	if data != nil {
		body, err = encodeBody(data)
		if err != nil {
			return err
		}
	} else {
		body = []byte{}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return NewClientError(resp)
	}

	return decodeBody(resp, out)
}

func decodeBody(resp *http.Response, v any) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

func encodeBody(data any) ([]byte, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	return b.Bytes(), err
}

var _ Connector = &HTTPClient{}
