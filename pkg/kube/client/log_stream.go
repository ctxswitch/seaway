package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	DefaultLogStreamerTimeout = 10 * time.Second
)

type objWriter struct {
	objRef corev1.ObjectReference
	writer io.Writer
}

// Write writes the provided bytes to the writer prefixed with the pods suffix.
func (ow *objWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	suffix := strings.Split(ow.objRef.Name, "-")
	prefix := []byte(fmt.Sprintf("[%s] ", suffix[len(suffix)-1]))

	n, err := ow.writer.Write(append(prefix, p...))
	if n > len(p) {
		return len(p), err
	}
	return n, err
}

type LogSteamer struct {
	client  *Client
	options corev1.PodLogOptions
}

// NewLogStreamer creates a new log streamer for the provided kube context and logging options.
func NewLogStreamer(kubeContext string, opts corev1.PodLogOptions) (*LogSteamer, error) {
	c, err := NewClient("", kubeContext)
	if err != nil {
		return nil, err
	}

	return &LogSteamer{
		client:  c,
		options: opts,
	}, nil
}

// PodLogs streams logs from the pods that match the provided labels.
func (ls *LogSteamer) PodLogs(ctx context.Context, ns, labels string) error {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)

	gv := scheme.Scheme.VersionsForGroupKind(schema.GroupKind{
		Group: "",
		Kind:  "PodList",
	})

	// Get the result object for the identified pod.  We use the label selector
	// so we can get any number of replicas that are running for the deployment.
	// Because of the label selector we'll always get a PodList object back.
	factory := ls.client.Factory()
	result := factory.NewBuilder().
		WithScheme(s, gv...).
		NamespaceParam(ns).DefaultNamespace().
		SingleResourceType().
		ResourceTypes("pods").
		LabelSelector(labels).Do()

	// Returns the list of resource infos that were found in the previous step.
	// Note to future self:  This represents the actual resource objects and not
	// the individual pods.  The pods are accessed through the resource object.
	infos, err := result.Infos()
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = fmt.Errorf("no resources found in %s namespace", ns)
		}
		return err
	}

	// If we don't have any pods listed in the resource object then bail.
	obj := infos[0].Object
	if len(obj.(*corev1.PodList).Items) == 0 {
		return fmt.Errorf("no resources found in %s namespace", ns)
	}

	return ls.stream(ctx, obj)
}

// stream initiates the log streaming process for the provided resource object.
func (ls *LogSteamer) stream(ctx context.Context, obj runtime.Object) error {
	requests, err := polymorphichelpers.LogsForObjectFn(ls.client, obj, &ls.options, DefaultLogStreamerTimeout, false)
	if err != nil {
		return err
	}

	if ls.options.Follow && len(requests) > 1 {
		return parallelConsumer(ctx, requests)
	}

	fmt.Println("sequentially")
	return sequentialConsumer(ctx, requests)
}

// parallelConsumer streams logs from multiple pods concurrently.  This borrows
// heavily from the kubectl codebase.
func parallelConsumer(ctx context.Context, requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	reader, writer := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(len(requests))
	for objRef, request := range requests {
		go func(objRef corev1.ObjectReference, request rest.ResponseWrapper) {
			defer wg.Done()
			out := &objWriter{
				objRef: objRef,
				writer: writer,
			}
			if err := consume(ctx, request, out); err != nil {
				// Ignore any errors.
				fmt.Fprintf(writer, "error: %v\n", err)
			}

		}(objRef, request)
	}

	go func() {
		wg.Wait()
		writer.Close()
	}()

	_, err := io.Copy(os.Stdout, reader)
	return err
}

// sequentialConsumer streams logs from multiple pods sequentially.  This borrows
// heavily from the kubectl codebase.
func sequentialConsumer(ctx context.Context, requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	for objRef, request := range requests {
		out := &objWriter{
			objRef: objRef,
			writer: os.Stdout,
		}
		if err := consume(ctx, request, out); err != nil {
			// Ignore any errors.
			fmt.Fprintf(os.Stdout, "error: %v\n", err)
		}
	}

	return nil
}

// consume reads the stream from the provided request and writes it to the provided writer.
// This also borrows heavily from the kubectl codebase.
func consume(ctx context.Context, request rest.ResponseWrapper, out io.Writer) error {
	readCloser, err := request.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	r := bufio.NewReader(readCloser)
	for {
		bytes, err := r.ReadBytes('\n')
		if _, err := out.Write(bytes); err != nil {
			return err
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}
