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

package client

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"ctx.sh/seaway/pkg/console"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Client defines a new kubernetes client which impletments the RESTClientGetter interface
// Originally the controller-runtime clients were used but as kustomize was introduced, more
// control was needed to handle the unstructured objects and interact with the API server.
// To make things consistent, all of the seactl clients will utilize this client.
type Client struct {
	config clientcmd.ClientConfig
}

// NewClient creates a new kubernetes client.  It takes a kubeconfig file and a kubeconfig
// named context as arguments.  Both valudes can be empty strings in which case the os
// default paths will be checked for the kubeconfig file and the default context will be
// used.
func NewClient(kubeconfig string, context string) (*Client, error) {
	config, err := loadConfig(kubeconfig, context)
	if err != nil {
		return nil, err
	}

	rawconfig, err := config.RawConfig()
	if err != nil {
		return nil, err
	}

	config = clientcmd.NewDefaultClientConfig(rawconfig, &clientcmd.ConfigOverrides{})

	return &Client{config}, nil
}

// Factory returns a new factory interface for the client.
func (c *Client) Factory() cmdutil.Factory {
	return cmdutil.NewFactory(c)
}

// ToDiscoveryClient returns a new discovery client interface for the client.
// Required for the RESTClientGetter interface.
func (c *Client) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	restconfig, err := c.config.ClientConfig()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(restconfig)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(dc), nil
}

// ToRESTMapper returns a new RESTMapper interface for the client. Required for the
// RESTClientGetter interface.
func (c *Client) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(dc), nil
}

// ToRawKubeConfigLoader returns a new clientcmd.ClientConfig interface for the client.
// Required for the RESTClientGetter interface.
func (c *Client) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.config
}

// ToRESTConfig returns a new rest.Config interface for the client. Required for the
// RESTClientGetter interface.
func (c *Client) ToRESTConfig() (*rest.Config, error) {
	return c.config.ClientConfig()
}

// ResourceInterfaceFor returns a new dynamic.ResourceInterface for the client.  It takes a
// namespace (can be an empty string) and a runtime.Object as arguments.  The runtime.Object
// is used to determine the group, version, and kind of the object which is then used to
// configure the resource interface.  We use the mapping to determine the scope of the object
// and if it is namespaced we create a namespaced resource interface.
func (c *Client) ResourceInterfaceFor(obj Object, method string) (dynamic.ResourceInterface, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()

	dc, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	dyn, err := c.Factory().DynamicClient()
	if err != nil {
		console.Fatal(err.Error())
	}

	var dr dynamic.ResourceInterface

	resource, err := ServerResourcesForGroupVersionKind(dc, gvk, method)
	if err != nil {
		return nil, err
	}

	gvr := gvk.GroupVersion().WithResource(resource.Name)
	if resource.Namespaced {
		if obj.GetNamespace() == "" {
			obj.SetNamespace(corev1.NamespaceDefault)
		}
		dr = dyn.Resource(gvr).Namespace(obj.GetNamespace())
	} else {
		dr = dyn.Resource(gvr)
	}

	return dr, nil
}

// ServerResourcesForGroupVersionKind returns the APIResource for the provided GroupVersionKind
func ServerResourcesForGroupVersionKind(dc discovery.DiscoveryInterface, gvk schema.GroupVersionKind, verb string) (*metav1.APIResource, error) {
	resources, err := dc.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}

	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			if supportedVerb(&r, verb) {
				return &r, nil
			}

			return nil, apierr.NewMethodNotSupported(
				schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
				verb,
			)
		}
	}

	return nil, apierr.NewNotFound(
		schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
		"",
	)
}

// supportedVerb returns true if the provided verb is supported by the APIResource
func supportedVerb(apiResource *metav1.APIResource, verb string) bool {
	if verb == "" || verb == "*" {
		return true
	}

	for _, v := range apiResource.Verbs {
		if strings.EqualFold(v, verb) {
			return true
		}
	}
	return false
}

// loadConfig loads the kubernetes configuration from the provided kubeconfig file, Borrowed
// heavily from the controller-runtime loader.
func loadConfig(kubeconfig, context string) (clientcmd.ClientConfig, error) {
	if kubeconfig != "" {
		return loadConfigWithContext(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			context,
		)
	}

	// TODO: Maybe support in-cluster configuration.  I don't think this will really be
	// necessary or wanted at this point.
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %w", err)
		}
		rules.Precedence = append(
			rules.Precedence,
			filepath.Join(u.HomeDir,
				clientcmd.RecommendedHomeDir,
				clientcmd.RecommendedFileName,
			))
	}

	return loadConfigWithContext(rules, context)
}

// loadConfigWithContext loads the kubernetes configuration from the provided loader and
// context.  Borrowed heavily from the controller-runtime loader.
func loadConfigWithContext(loader clientcmd.ClientConfigLoader, context string) (clientcmd.ClientConfig, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}), nil
}
