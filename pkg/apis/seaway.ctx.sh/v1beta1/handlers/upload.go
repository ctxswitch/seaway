package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:skip
type UploadOptions struct {
	Client        client.Client
	Namespace     string
	StorageURL    string
	StorageBucket string
	StoragePrefix string
	StorageRegion string
}

// +kubebuilder:skip
type Upload struct {
	options *UploadOptions
	client.Client
}

func NewUploadHandler(opts *UploadOptions) http.Handler {
	if opts.Namespace == "" {
		opts.Namespace = v1beta1.DefaultControllerNamespace
	}

	upload := &Upload{
		options: opts,
		Client:  opts.Client,
	}

	return http.HandlerFunc(upload.ServeHTTP)
}

func (h *Upload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	logger.Info("Received a file upload request")
	_ = r.ParseMultipartForm(200 << 20) // 200 MB
	name := r.FormValue("name")
	namespace := r.FormValue("namespace")
	etag := r.FormValue("etag")
	config := r.FormValue("config")

	if config == "" {
		h.respond(w, minio.UploadInfo{}, fmt.Errorf("config was not specified"))
	}

	parts := strings.Split(config, "/")
	if len(parts) == 1 {
		parts = append([]string{h.options.Namespace}, parts...)
	}

	logger.Info("Uploading file", "name", name, "namespace", namespace, "etag", etag, "config", config)

	file, _, err := r.FormFile("archive")
	if err != nil {
		logger.Error(err, "Error retrieving the file", "name", name, "namespace", namespace, "etag", etag)
		h.respond(w, minio.UploadInfo{}, err)
		return
	}

	info, err := h.store(ctx, file, name, namespace)
	if err != nil {
		logger.Error(err, "Error storing the file", "name", name, "namespace", namespace, "etag", etag)
		h.respond(w, minio.UploadInfo{}, err)
		return
	}

	if etag != info.ETag {
		logger.Info("ETag mismatch", "name", name, "namespace", namespace, "src", etag, "dst", info.ETag)
		h.respond(w, info, fmt.Errorf("ETag mismatch"))
		return
	}

	logger.Info("File uploaded", "name", name, "namespace", namespace, "etag", etag, "file", info.Key, "size", info.Size)
	h.respond(w, info, nil)
}

func (h *Upload) store(ctx context.Context, file multipart.File, name, namespace string) (minio.UploadInfo, error) {
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return minio.UploadInfo{}, err
	}

	// Split these out so we can release mem earlier.
	tmpkey := fmt.Sprintf("upload-%s-%s.tar.gz", name, namespace)
	tmp, err := os.CreateTemp("", tmpkey)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	defer tmp.Close()

	// TODO: Support more providers than the environment providers.  There's
	// others like IAM, STS, etc.  We could also add a secrets provider as well.
	u, err := url.Parse(h.options.StorageURL)
	if err != nil {
		return minio.UploadInfo{}, err
	}

	// TODO: Chicken/egg
	store, err := minio.New(u.Host, &minio.Options{
		Creds: credentials.NewChainCredentials([]credentials.Provider{
			// Requires MINIO_ACCESS_KEY and MINIO_SECRET_KEY.
			&credentials.EnvMinio{},
			// Requires AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.
			&credentials.EnvAWS{},
		}),
		Region: h.options.StorageRegion,
	})
	if err != nil {
		return minio.UploadInfo{}, err
	}

	err = h.makeBucketIfNotExists(ctx, store, h.options.StorageBucket)
	if err != nil {
		return minio.UploadInfo{}, err
	}

	_, err = tmp.Write(data)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	archive := tmp.Name()

	// Upload to minio.
	archiveKey := fmt.Sprintf("%s/%s-%s.tar.gz", h.options.StoragePrefix, name, namespace)
	info, err := store.FPutObject(ctx, h.options.StorageBucket, archiveKey, archive, minio.PutObjectOptions{})
	if err != nil {
		return minio.UploadInfo{}, err
	}

	return info, nil
}

func (h *Upload) makeBucketIfNotExists(ctx context.Context, store *minio.Client, bucket string) error {
	ok, err := store.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	return store.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
		ObjectLocking: false,
		Region:        "",
	})
}

// Respond with either the file info or an error.
func (h *Upload) respond(w http.ResponseWriter, info minio.UploadInfo, err error) {
	if err != nil {
		resp := v1beta1.UploadResponse{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := v1beta1.UploadResponse{
		Key:  info.Key,
		ETag: info.ETag,
		Size: info.Size,
		Code: http.StatusOK,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

var _ http.Handler = &Upload{}
