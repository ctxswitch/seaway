package env

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"ctx.sh/seaway/pkg/storage"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	SyncUsage     = "sync [context]"
	SyncShortDesc = "Sync to the target object storage using the configuration context"
	SyncLongDesc  = `Sync the code to the target object storage based on the configuration context
provided in the manifest.  This will trigger a new development deployment if there was a change.`
)

type Sync struct {
	logLevel int8
}

func NewSync() *Sync {
	return &Sync{}
}

func (s *Sync) RunE(cmd *cobra.Command, args []string) error {
	ctx := ctrl.SetupSignalHandler()

	if len(args) != 1 {
		return fmt.Errorf("expected context name")
	}

	var manifest v1beta1.Manifest
	if err := manifest.Load("manifest.yaml"); err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build context '%s' not found in the manifest", args[0])
	}

	store := storage.NewClient(env.Source.S3.GetEndpoint(), env.Source.S3.UseSSL())
	mc, err := store.Connect(ctx, nil)
	if err != nil {
		console.Fatal("Unable to connect to object storage: %s", err)
	}

	console.Info("Creating archive")
	archive, err := s.create(manifest.Name, &env)
	if err != nil {
		console.Fatal("Unable to create archive: %s", err)
	}
	defer os.Remove(archive)

	console.Info("Uploading archive")
	bucket := *env.Source.S3.Bucket
	key := fmt.Sprintf("%s/%s-%s.tar.gz", *env.Source.S3.Prefix, manifest.Name, env.Namespace)

	info, err := mc.FPutObject(ctx, bucket, key, archive, minio.PutObjectOptions{})
	if err != nil {
		console.Fatal("Unable to upload the archive: %s", err)
	}

	console.Notice("Source: %s", archive)
	console.Notice("Destination: s3://%s/%s", bucket, key)
	console.Notice("Size: %d", info.Size)
	console.Notice("ETag: %s", info.ETag)

	console.Info("Updating environment")
	client, err := kube.NewSeawayClient("", "")
	if err != nil {
		console.Fatal("error getting seaway client: %s", err.Error())
	}

	obj := GetEnvironment(manifest.Name, env.Namespace)

	op, err := client.CreateOrUpdate(ctx, obj, func() error {
		env.EnvironmentSpec.DeepCopyInto(&obj.Spec)
		obj.Spec.Revision = info.ETag
		return nil
	})
	if err != nil {
		console.Fatal("error modifying environment: %s", err.Error())
	}

	switch op {
	case kube.OperationResultNone:
		console.Success("No changes detected")
		return nil
	case kube.OperationResultUpdated:
		console.Info("Environment updated")
	case kube.OperationResultCreated:
		console.Info("Environment created")
	}

	// TODO: timeout should be configurable
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	status := ""

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			console.Warn("Cancelled: %s", ctx.Err())
			return nil
		case <-ticker.C:
			// TODO: Best to use a clean environment object here?
			if err := client.Get(ctx, obj, metav1.GetOptions{}); err != nil {
				console.Fatal("Unable to get the environment: %s", err)
			}
			if obj.IsDeployed() {
				console.Success("Revision has been deployed")
				return nil
			} else if status != string(obj.Status.Stage) {
				status = string(obj.Status.Stage)
				console.Notice(status)
			}
		}
	}
}

func (s *Sync) create(name string, env *v1beta1.ManifestEnvironmentSpec) (string, error) {
	out, err := os.CreateTemp("", name+"-*.tar.gz")
	if err != nil {
		console.Fatal("Unable to create the temporary archive: %s", err)
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	includes := env.Includes()
	excludes := env.Excludes()

	filepath.WalkDir(".", func(f string, d fs.DirEntry, e error) error {
		include := includes.MatchString(f)
		exclude := excludes.MatchString(f)
		if include && !exclude {
			console.ListItem(f)
			s.add(tw, f)
		}

		return nil
	})

	return out.Name(), nil
}

func (s *Sync) add(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = filename
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sync) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   SyncUsage,
		Short: SyncShortDesc,
		Long:  SyncLongDesc,
		RunE:  s.RunE,
	}

	cmd.PersistentFlags().Int8VarP(&s.logLevel, "log-level", "", DefaultLogLevel, "set the log level (integer value)")

	return cmd
}

func GetEnvironment(name, namespace string) *v1beta1.Environment {
	env := &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	// TODO: We actually need this with the client at this point because we use
	// the gvk to get the resource interface. Revisit this later and refactor it
	// out.  It's not horrible but it's not great either.
	env.SetGroupVersionKind(v1beta1.SchemeGroupVersion.WithKind("Environment"))

	return env
}
