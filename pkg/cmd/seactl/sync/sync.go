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

package sync

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/md5" //nolint:gosec
	"ctx.sh/seaway/pkg/cmd/util"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	kube "ctx.sh/seaway/pkg/kube/client"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultTimeout = 10 * time.Minute
)

type Command struct {
	OnlyDeps bool
	WithDeps bool
	LogLevel int8
	Force    bool
}

// RunE is the main function for the sync command which syncs the code to the target
// object storage and creates or updates the development environment.
// TODO: address the linting issues.
func (c *Command) RunE(cmd *cobra.Command, args []string) error { //nolint:funlen,gocognit
	kubeContext := cmd.Root().Flags().Lookup("context").Value.String()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(args) != 1 {
		return fmt.Errorf("expected environment name")
	}

	var manifest v1beta1.Manifest
	err := manifest.Load("manifest.yaml")
	if err != nil {
		console.Fatal("Unable to load manifest")
	}

	env, err := manifest.GetEnvironment(args[0])
	if err != nil {
		console.Fatal("Build context '%s' not found in the manifest", args[0])
	}

	client, err := kube.NewKubectlCmd("", kubeContext)
	if err != nil {
		console.Fatal(err.Error())
	}

	if c.WithDeps || c.OnlyDeps {
		err := doApply(ctx, client, env)
		if c.OnlyDeps || err != nil {
			return err
		}
	}

	return doSync(ctx, client, manifest.Name, env, c.Force)
}

func doApply(ctx context.Context, client *kube.KubectlCmd, env v1beta1.ManifestEnvironmentSpec) error {
	return apply(ctx, client, env)
}

//nolint:funlen,gocognit
func doSync(ctx context.Context, client *kube.KubectlCmd, name string, env v1beta1.ManifestEnvironmentSpec, force bool) error {
	console.Info("Creating archive")
	archive, err := create(name, env)
	if err != nil {
		console.Fatal("Unable to create archive: %s", err)
	}
	defer func() {
		_ = os.Remove(archive)
	}()

	console.Info("Uploading archive")

	etag, err := checksum(archive)
	if err != nil {
		console.Fatal("Unable to calculate the archive checksum: %s", err)
	}

	upload := v1beta1.NewClient(env.Endpoint + "/upload")
	resp, err := upload.Upload(ctx, archive, map[string]string{
		"name":      name,
		"namespace": env.Namespace,
		"etag":      etag,
		"config":    env.Name,
	})
	if err != nil {
		console.Fatal("Unable to upload the archive: %v", err)
	}

	if resp.Code != 200 {
		console.Fatal("Upload failed: %s", resp.Error)
	}

	// console.Notice("Source: %s", archive)
	console.ListNotice("Size: %d", resp.Size)
	console.ListNotice("Revision: %s", resp.ETag)

	console.Info("Deploying")

	obj := util.GetEnvironment(name, env.Namespace)

	if force {
		derr := client.Delete(ctx, obj, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(derr) {
			console.Fatal("error deleting environment: %s", derr.Error())
		}
	}

	op, err := client.CreateOrUpdate(ctx, obj, func() error {
		env.EnvironmentSpec.DeepCopyInto(&obj.Spec)
		obj.Spec.Revision = resp.ETag
		obj.Spec.Config = env.Name
		return nil
	})
	if err != nil {
		console.Fatal("error modifying environment: %s", err.Error())
	}

	switch op {
	case kube.OperationResultNone:
		console.ListNotice("No changes detected")
		return nil
	case kube.OperationResultUpdated:
		console.ListNotice("Environment updated")
	case kube.OperationResultCreated:
		console.ListNotice("Environment created")
	}

	// TODO: timeout should be configurable
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
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
			err = client.Get(ctx, obj, metav1.GetOptions{})
			if err != nil {
				console.Fatal("Unable to get the environment: %s", err)
			}

			// TODO: fix me.  don't convert.
			if status != string(obj.Status.Stage) {
				status = string(obj.Status.Stage)
				switch {
				case obj.IsFailing():
					console.ListWarning(status)
				case obj.HasFailed():
					console.ListFailed(status)
					return nil
				case obj.IsDeployed():
					console.ListSuccess(status)
					return nil
				default:
					console.ListNotice(status)
				}
			}
		}
	}
}

// create builds the tar/gzip archive that will be uploaded to the object storage.
func create(name string, env v1beta1.ManifestEnvironmentSpec) (string, error) {
	out, err := os.CreateTemp("", name+"-*.tar.gz")
	if err != nil {
		console.Fatal("Unable to create the temporary archive: %s", err)
	}
	defer func() {
		_ = out.Close()
	}()

	gw := gzip.NewWriter(out)
	defer func() {
		_ = gw.Close()
	}()
	tw := tar.NewWriter(gw)
	defer func() {
		_ = tw.Close()
	}()

	includes := env.Includes()
	excludes := env.Excludes()

	err = filepath.WalkDir(".", func(f string, d fs.DirEntry, e error) error {
		include := includes.MatchString(f)
		exclude := excludes.MatchString(f)
		if include && !exclude {
			console.ListItem(f)
			if aerr := add(tw, f); aerr != nil {
				return aerr
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}

// add adds a file to the archive.
func add(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
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

func checksum(filename string) (string, error) {
	h := md5.New() //nolint:gosec
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
