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

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"ctx.sh/seaway/pkg/console"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
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
type Client struct {
	config clientcmd.ClientConfig
}

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

func (c *Client) Factory() cmdutil.Factory {
	return cmdutil.NewFactory(c)
}

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

func (c *Client) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(dc), nil
}

func (c *Client) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.config
}

func (c *Client) ToRESTConfig() (*rest.Config, error) {
	return c.config.ClientConfig()
}

func (c *Client) EnvironmentInterface(ns string) (dynamic.ResourceInterface, error) {
	dyn, err := c.Factory().DynamicClient()
	if err != nil {
		console.Fatal(err.Error())
	}

	return dyn.Resource(schema.GroupVersionResource{
		Group:    v1beta1.SchemeGroupVersion.Group,
		Version:  v1beta1.SchemeGroupVersion.Version,
		Resource: "environments",
	}).Namespace(ns), nil
}

// func (c *Client) ResourceFor(obj runtime.Object) (schema.GroupVersionResource, error) {
// 	gvk := obj.GetObjectKind().GroupVersionKind()

// 	dc, err := c.ToDiscoveryClient()
// 	if err != nil {
// 		return schema.GroupVersionResource{}, err
// 	}
// 	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

// 	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
// 	if err != nil {
// 		return schema.GroupVersionResource{}, err
// 	}

// 	return mapping.Resource, nil
// }

func (c *Client) ResourceInterfaceFor(ns string, obj runtime.Object) (dynamic.ResourceInterface, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()

	dc, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		console.Fatal(err.Error())
	}

	dyn, err := c.Factory().DynamicClient()
	if err != nil {
		console.Fatal(err.Error())
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = dyn.Resource(mapping.Resource).Namespace(ns)
	} else {
		dr = dyn.Resource(mapping.Resource)
	}

	return dr, nil
}

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

func loadConfigWithContext(loader clientcmd.ClientConfigLoader, context string) (clientcmd.ClientConfig, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}), nil
}
