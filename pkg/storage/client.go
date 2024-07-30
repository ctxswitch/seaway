package storage

import (
	"errors"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	Endpoint string
	UseSSL   bool
}

func NewClient(endpoint string, useSSL bool) *Client {
	return &Client{
		Endpoint: endpoint,
		UseSSL:   useSSL,
	}
}

func (c *Client) Connect() (*minio.Client, error) {
	id := os.Getenv("SEAWAY_ACCESS_KEY")
	if id == "" {
		return nil, errors.New("SEAWAY_ACCESS_KEY not set")
	}

	key := os.Getenv("SEAWAY_SECRET_KEY")
	if key == "" {
		return nil, errors.New("SEAWAY_SECRET_KEY not set")
	}

	// Initialize minio client object.
	mc, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(id, key, ""),
		Secure: c.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return mc, nil
}
