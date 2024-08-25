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

package storage

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Client struct {
	Endpoint string
	UseSSL   bool
}

func NewClient(endpoint string, useSSL bool) *Client {
	return &Client{
		Endpoint: endpoint,
		UseSSL:   useSSL,
		// TODO: Move creds to here and don't process them in the connect functions.
	}
}

func (c *Client) Connect(ctx context.Context, secret *corev1.Secret) (*minio.Client, error) {
	logger := log.FromContext(ctx)

	if secret != nil {
		// TODO: Fall back to env vars if there is an error getting the secret?
		return c.connectWithSecret(secret)
	}

	logger.Info("using env vars for s3 credentials")
	return c.connectWithEnv()
}

func (c *Client) connectWithEnv() (*minio.Client, error) {
	id := os.Getenv("AWS_ACCESS_KEY_ID")
	if id == "" {
		return nil, errors.New("AWS_ACCESS_KEY_ID not set")
	}

	key := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if key == "" {
		return nil, errors.New("AWS_SECRET_ACCESS_KEY not set")
	}

	// Initialize minio client object.
	return minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(id, key, ""),
		Secure: c.UseSSL,
	})
}

// TODO: I have the data fields differently named here, so update env to match.  I'm using those
// because that is what minio expects.  Need to look at this more.
func (c *Client) connectWithSecret(secret *corev1.Secret) (*minio.Client, error) {
	id, ok := secret.Data["AWS_ACCESS_KEY_ID"]
	if !ok {
		return nil, errors.New("AWS_ACCESS_KEY_ID not found in secret")
	}

	key, ok := secret.Data["AWS_SECRET_ACCESS_KEY"]
	if !ok {
		return nil, errors.New("AWS_SECRET_ACCESS_KEY not found in secret")
	}

	idText := strings.Trim(string(id), "\n")
	keyText := strings.Trim(string(key), "\n")

	// Initialize minio client object.
	return minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(idText, keyText, ""),
		Secure: c.UseSSL,
	})
}
