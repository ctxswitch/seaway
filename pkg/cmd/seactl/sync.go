package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"
	"stvz.io/seaway/pkg/core"
	"stvz.io/seaway/pkg/storage"
	"stvz.io/seaway/pkg/utils"
)

const (
	SyncUsage     = "sync"
	SyncShortDesc = "Sync to the target object storage"
	SyncLongDesc  = `Sync the code to the target object storage based on the configuration 
provided in the manifest.  This will trigger a new development deployment.`
)

type Sync struct{}

func NewSync() *Sync {
	return &Sync{}
}

func (s *Sync) RunE(cmd *cobra.Command, args []string) error {
	manifest := core.NewManifest()
	if err := manifest.Load("manifest.yaml"); err != nil {
		log.Fatalln(err)
	}

	store := storage.NewClient(manifest.Seaway.Endpoint, manifest.Seaway.UseSSL)
	mc, err := store.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	includes := manifest.Includes()
	excludes := manifest.Excludes()

	// Limit this to the current directory and get the includes from the manifest
	// If you exclude the manifest, you won't be able to build/test/run
	sums := make(map[string]string)

	filepath.WalkDir(".", func(f string, d fs.DirEntry, e error) error {
		info, err := os.Stat(f)
		if err != nil {
			log.Fatalln("stat", err)
		}

		if !info.IsDir() && includes.MatchString(f) && !excludes.MatchString(f) {
			sums[f] = utils.ETag(f)
		}

		return nil
	})

	sumfile, err := os.Create("seaway.sum")
	if err != nil {
		log.Fatalln(err)
	}

	for k, v := range sums {
		fmt.Fprintf(sumfile, "%s %s\n", k, v)
	}

	for file, sum := range sums {
		key := manifest.Name + "/" + file

		objectInfo, err := mc.StatObject(context.Background(), "development", key, minio.StatObjectOptions{})
		if err != nil {
			log.Println(err)
		}

		if objectInfo.ETag != sum {
			log.Printf("Uploading %s\n", file)
			_, err = mc.FPutObject(context.Background(), "development", key, file, minio.PutObjectOptions{})
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	for file := range mc.ListObjects(context.Background(), "development", minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    manifest.Name + "/",
	}) {
		if file.Err != nil {
			log.Fatalln(file.Err)
		}

		// Strip the prefix
		file.Key = file.Key[len(manifest.Name)+1:]

		if _, ok := sums[file.Key]; !ok && file.Key != ".seaway.metadata" {
			// pull the current metadata
			log.Printf("Deleting %s\n", file.Key)
			err = mc.RemoveObject(context.Background(), "development", file.Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Fatalln(err)
			}
		}

	}

	metadata := core.NewMetadata(sums)
	data, err := metadata.Marshal()
	if err != nil {
		log.Fatalln(err)
	}

	key := "_metadata/" + manifest.Name + "/metadata.json"
	_, err = mc.PutObject(context.Background(), "development", key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		log.Fatalln(err)
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

	return cmd
}
