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

package kustomize

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

// ResourceKey is a struct that defines a unique map key to the resources.  It's
// used to quickly access specific resources from parsed documents.
// TODO: Might need to add namespace to this. May need APIVersion to distinguish kinds
// with same names.  For right now, this will do.
type ResourceKey struct {
	Name string
	Kind string
}

// KustomizerResource is a struct that contains an unstructured object and its
// GroupVersionKind.
type KustomizerResource struct {
	Resource *unstructured.Unstructured
	GVK      *schema.GroupVersionKind
}

// KustomizerOptions contains the configuration options for the Kustomizer.
// TODO: There are several more options that may be useful to add here.
type KustomizerOptions struct {
	BaseDir string
}

// Kustomizer processes a kustomize directory and returns the generated
// resources.
type Kustomizer struct {
	raw  []byte
	docs *utilyaml.YAMLReader

	order       []ResourceKey
	resourceMap map[ResourceKey]KustomizerResource
}

// NewKustomizer creates, configures, and runs kustomize on the specified
// directory.
func NewKustomizer(opts *KustomizerOptions) (*Kustomizer, error) {
	kustomizer := krusty.MakeKustomizer(&krusty.Options{
		Reorder:           "none",
		AddManagedbyLabel: false,
		LoadRestrictions:  types.LoadRestrictionsRootOnly,
		PluginConfig: &types.PluginConfig{
			HelmConfig: types.HelmConfig{
				Enabled: true,
				Command: "helm",
			},
		},
	})

	target := filesys.MakeFsOnDisk()

	r, err := kustomizer.Run(target, opts.BaseDir)
	if err != nil {
		return nil, err
	}

	yml, err := r.AsYaml()
	if err != nil {
		return nil, err
	}

	return &Kustomizer{
		raw:         yml,
		docs:        utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(yml))),
		order:       make([]ResourceKey, 0),
		resourceMap: make(map[ResourceKey]KustomizerResource),
	}, nil
}

// NewKustomizerFromBytes rehydrates previously kustomized raw bytes.
func NewKustomizerFromBytes(raw []byte) (*Kustomizer, error) {
	return &Kustomizer{
		raw:         raw,
		docs:        utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(raw))),
		order:       make([]ResourceKey, 0),
		resourceMap: make(map[ResourceKey]KustomizerResource),
	}, nil
}

func (k *Kustomizer) Build() error {
	for {
		doc, err := k.docs.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		// TODO: envsubst here or another template type maybe `{{ var }}`.

		decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		decoded := &unstructured.Unstructured{}

		_, gvk, err := decoder.Decode(doc, nil, decoded)
		if err != nil {
			return err
		}

		resource := KustomizerResource{
			Resource: decoded,
			GVK:      gvk,
		}

		key := ResourceKey{
			Name: resource.Resource.GetName(),
			Kind: resource.GVK.GroupKind().Kind,
		}

		// TODO: clean me up.
		k.resourceMap[key] = resource
		k.order = append(k.order, key)
	}

	return nil
}

// Resources returns a list of resources in the order they were parsed by kustomize.
func (k *Kustomizer) Resources() []KustomizerResource {
	resources := make([]KustomizerResource, 0)
	for _, key := range k.order {
		resources = append(resources, k.resourceMap[key])
	}

	return resources
}

func (k *Kustomizer) ResourceMap() map[ResourceKey]KustomizerResource {
	return k.resourceMap
}

// GetResource returns a single resource from the resource map.
func (k *Kustomizer) GetResource(kind, name string) (res KustomizerResource, ok bool) {
	res, ok = k.resourceMap[ResourceKey{Name: name, Kind: kind}]
	return
}

// SetResource modifies an existing resource from the map.  The resource is required to exist
// to preserve any ordering that kustomize has set.
func (k *Kustomizer) SetResource(resource *unstructured.Unstructured) error {
	key := ResourceKey{Name: resource.GetName(), Kind: resource.GetKind()}
	res, ok := k.resourceMap[key]
	if !ok {
		return fmt.Errorf("unable to add resource: %s", resource.GetName())
	}

	k.resourceMap[key] = KustomizerResource{
		Resource: resource,
		GVK:      res.GVK,
	}

	return nil
}

// Raw returns the raw yaml bytes that were generated with kustomize.
func (k *Kustomizer) Raw() []byte {
	return k.raw
}

// ToYamlBytes converts the kustomized objects back to their yaml form.  This is a temporary
// solution to perform substitutions before the apply.
func (k *Kustomizer) ToYamlBytes() ([]byte, error) {
	resources := make([]string, 0)
	for _, key := range k.order {
		resource := k.resourceMap[key].Resource

		y, err := yamlv3.Marshal(resource.Object)
		if err != nil {
			return nil, err
		}

		resources = append(resources, string(y))
	}

	return []byte(strings.Join(resources, "---\n")), nil
}
