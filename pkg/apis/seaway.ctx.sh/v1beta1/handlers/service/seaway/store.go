package seaway

import (
	"context"
	"io"
	"net/url"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StoreOptions struct {
	Endpoint string
	Region   string
	Bucket   string
}

type Store struct {
	options *StoreOptions
	reader  *io.PipeReader
	writer  *io.PipeWriter
	info    minio.UploadInfo
	err     error
	done    chan struct{}
	putOnce sync.Once
	sync.Mutex
}

func NewStore(options *StoreOptions) *Store {
	r, w := io.Pipe()
	return &Store{
		options: options,
		writer:  w,
		reader:  r,
		done:    make(chan struct{}),
	}
}

func (s *Store) Put(ctx context.Context, key string) {
	defer close(s.done)

	s.putOnce.Do(func() {
		store, err := s.getStore()
		if err != nil {
			s.err = err
			return
		}

		bucket := s.options.Bucket
		info, err := store.PutObject(ctx, bucket, key, s.reader, -1, minio.PutObjectOptions{})
		if err != nil {
			s.err = err
			return
		}

		s.info = info
	})
}

func (s *Store) EnsureBucket(ctx context.Context, bucket string) error {
	store, err := s.getStore()
	if err != nil {
		return err
	}

	ok, err := store.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return store.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
		Region: s.options.Region,
	})
}

func (s *Store) Info() minio.UploadInfo {
	return s.info
}

func (s *Store) Write(b []byte) error {
	s.Lock()
	defer s.Unlock()
	_, err := s.writer.Write(b)
	if err != nil {
		s.CloseWithError(err)
		return err
	}

	return nil
}

func (s *Store) Wait() {
	<-s.done
	s.Close()
}

func (s *Store) Err() error {
	return s.err
}

func (s *Store) Close() {
	_ = s.writer.Close()
}

func (s *Store) CloseWithError(err error) {
	_ = s.writer.CloseWithError(err)
}

func (s *Store) getStore() (*minio.Client, error) {
	endpoint, err := url.Parse(s.options.Endpoint)
	if err != nil {
		return nil, err
	}

	return minio.New(endpoint.Host, &minio.Options{
		Creds: credentials.NewChainCredentials([]credentials.Provider{
			// Requires MINIO_ACCESS_KEY and MINIO_SECRET_KEY.
			&credentials.EnvMinio{},
			// Requires AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.
			&credentials.EnvAWS{},
		}),
		Region: s.options.Region,
	})
}
