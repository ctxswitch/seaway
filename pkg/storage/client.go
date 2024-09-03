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

	"ctx.sh/seaway/pkg/auth"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client is a wrapper around the minio client.
type Client struct {
	Endpoint string
	UseSSL   bool

	client *minio.Client
}

// NewClient creates a new client.
func NewClient(endpoint string, useSSL bool) *Client {
	return &Client{
		Endpoint: endpoint,
		UseSSL:   useSSL,
		// TODO: Move creds to here and don't process them in the connect functions.
	}
}

// Connect creates a connection to the S3 storage service.  It's only really
// been tested against Minio, but in theory should work with any S3-compatible
// service as long as the configuration is correct.
func (c *Client) Connect(ctx context.Context, creds *auth.Credentials) error {
	client, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(creds.GetAccessKey(), creds.GetSecretKey(), ""),
		Secure: c.UseSSL,
		// TODO: Add region support.
		// TODO: Add secure session support.
		// TODO: Potentially add trace client support.
		// TODO: Add bucket lookup support.
		// TODO: Add trailing headers support.
	})
	if err != nil {
		return err
	}
	c.client = client

	return nil
}

// CreateBucketIfNotExists creates a bucket if it doesn't already exist.
func (c *Client) CreateBucketIfNotExists(ctx context.Context, client *minio.Client, bucket string) error {
	ok, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
		ObjectLocking: false,
		Region:        "",
	})
}

func (c *Client) PutObject(ctx context.Context, bucket, key string, file string) (minio.UploadInfo, error) {
	err := c.CreateBucketIfNotExists(ctx, c.client, bucket)
	if err != nil {
		return minio.UploadInfo{}, err
	}

	return c.client.FPutObject(ctx, bucket, key, file, minio.PutObjectOptions{})
}

func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	return c.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

// info, err := mc.FPutObject(ctx, bucket, key, archive, minio.PutObjectOptions{})
// 	if err != nil {
// 		console.Fatal("Unable to upload the archive: %s", err)
// 	}
