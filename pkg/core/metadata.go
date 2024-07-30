package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/minio/minio-go/v7"
)

type Metadata struct {
	Sums map[string]string `json:"files"`
}

func NewMetadata(sums map[string]string) *Metadata {
	return &Metadata{
		Sums: sums,
	}
}

func (m *Metadata) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (m *Metadata) Lock(mc *minio.Client, manifest *Manifest) (bool, error) {
	mc.PutObject(context.TODO(), "development", MetadataLockPath(manifest), nil, 0, minio.PutObjectOptions{
		// Expire the lock after 5 minutes in case the release fails.
		Expires: time.Now().Add(5 * time.Minute),
	})
}

func Unmarshal(data []byte) (*Metadata, error) {
	m := &Metadata{}
	err := json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
