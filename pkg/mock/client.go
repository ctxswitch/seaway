package mock

import (
	"os"
	"path/filepath"
	"strings"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Client struct {
	fixtureDir string
	log        logr.Logger
	tracker    testing.ObjectTracker
	scheme     *runtime.Scheme
	client.Client
}

// NewClient returns a mock (fake) client for testing. The fixtures are
// not automatically loaded into the cache.  Individual fixtures can be loaded
// using the ApplyFixtureOrDie method.
func NewClient() *Client {
	s := scheme.Scheme
	_ = v1beta1.AddToScheme(s)

	tracker := testing.NewObjectTracker(s, scheme.Codecs.UniversalDecoder())
	client := fake.NewClientBuilder().
		WithObjectTracker(tracker).
		WithScheme(s).
		WithStatusSubresource(&v1beta1.Environment{}).
		Build()

	return &Client{
		fixtureDir: "",
		log:        logr.Discard(),
		scheme:     s,
		tracker:    tracker,
		Client:     client,
	}
}

// WithLogger sets the logger for the client.
func (m *Client) WithLogger(log logr.Logger) *Client {
	m.log = log
	return m
}

func (m *Client) WithFixtureDirectory(dir string) *Client {
	m.fixtureDir = dir
	return m
}

// ApplyFixtureOrDie loads a single fixture into the cache.  The fixture must be in a
// recognizable format for the universal deserializer.
func (m *Client) ApplyFixtureOrDie(filename ...string) {
	decoder := scheme.Codecs.UniversalDeserializer()
	for _, file := range filename {
		f := filepath.Join(m.fixtureDir, file)
		data, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}

		sections := strings.Split(string(data), "---")

		for _, section := range sections {
			data = []byte(section)
			obj, _, err := decoder.Decode(data, nil, nil)
			if err != nil {
				panic(err)
			}

			// Fake some of the creation metadata.  There's probably a few other
			// things that could be useful.
			if obj.(client.Object).GetCreationTimestamp().Time.IsZero() {
				obj.(client.Object).SetCreationTimestamp(metav1.Time{
					Time: metav1.Now().Time,
				})
			}

			// If the namespace is not set, set it to default.
			if obj.(client.Object).GetNamespace() == "" {
				obj.(client.Object).SetNamespace("default")
			}

			err = m.tracker.Add(obj)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (m *Client) Reset() {
	m.tracker = testing.NewObjectTracker(m.scheme, scheme.Codecs.UniversalDecoder())
}

var _ client.Client = &Client{}
