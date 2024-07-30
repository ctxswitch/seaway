package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"stvz.io/seaway/pkg/core"
	"stvz.io/seaway/pkg/storage"
)

const (
	ListenerUsage     = "listener"
	ListenerShortDesc = "Listen for events"
	ListenerLongDesc  = `Listens for sync events and handles any updates`
)

type Listener struct {
	Metadata map[string]*core.Metadata
}

func NewListener() *Listener {
	return &Listener{}
}

func (l *Listener) RunE(cmd *cobra.Command, args []string) error {
	manifest := core.NewManifest()
	if err := manifest.Load("manifest.yaml"); err != nil {
		log.Fatalln(err)
	}

	store := storage.NewClient(manifest.Seaway.Endpoint, manifest.Seaway.UseSSL)
	mc, err := store.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	events := []string{
		"s3:ObjectCreated:Put",
		"s3:ObjectRemoved:Delete",
	}

	for evt := range mc.ListenBucketNotification(context.TODO(), "development", "_metadata/", "", events) {
		if evt.Err != nil {
			log.Println("bad event", evt.Err)
		}
		for _, record := range evt.Records {
			// How do I track/reconcile in the case of a disruption? Do I need to keep track of the
			// metadata elsewhere and only clean up the metadata when the objects have been finalized?
			// i.e. the listener will listen for the metadata update, if pending deletion, then delete
			// the deployment, delete the source target, and then delete the metadata.
			// if record.EventName == "s3:ObjectRemoved:Put" {

			// } else if record.EventName == "s3:ObjectRemoved:Delete" {

			// 	delete(l.Metadata, record.S3.Object.Key)
			// 	continue
			// }

			// obj, err := mc.GetObject(context.TODO(), "development", record.S3.Object.Key, minio.GetObjectOptions{})
			// if err != nil {
			// 	log.Println("get object", err)
			// }

			// data := make([]byte, record.S3.Object.Size)
			// _, err = obj.Read(data)
			// if err != nil {
			// 	log.Println("read object", err)
			// }

			// metadata, err := core.Unmarshal(data)
			// if err != nil {
			// 	log.Println("unmarshal", err)
			// }

			// l.Metadata[record.S3.Object.Key] = metadata

			// and then walk the S3 prefix comparing the ETags.  The moment we find a difference
			// we store all the sums and send a restart to the development pod/deployment.
			log.Printf("[%s] %s %s\n", record.EventName, record.S3.Object.Key, record.S3.Object.ETag)
		}
	}
	return nil
}

func (l *Listener) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ListenerUsage,
		Short: ListenerShortDesc,
		Long:  ListenerLongDesc,
		RunE:  l.RunE,
	}

	return cmd
}
