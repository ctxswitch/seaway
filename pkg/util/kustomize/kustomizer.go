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
	docs *utilyaml.YAMLReader

	// TODO: consolidate these down to a list of strings and the map with the
	// resources.  Turn into a generator for the apply.
	resources   []KustomizerResource
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
		docs:        utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(yml))),
		resources:   make([]KustomizerResource, 0),
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

		name := resource.Resource.GetName()
		kind := resource.GVK.GroupKind().Kind

		// TODO: clean me up.
		k.resourceMap[ResourceKey{Name: name, Kind: kind}] = resource
		k.resources = append(k.resources, resource)
	}

	return nil
}

// Resources reads through all of the generated documents and decodes them into
// unstructured objects.  It returns a list of KustomizerResource objects containing
// the unstructured object and its GroupVersionKind.
func (k *Kustomizer) Resources() []KustomizerResource {
	return k.resources
}

func (k *Kustomizer) ResourceMap() map[ResourceKey]KustomizerResource {
	return k.resourceMap
}

func (k *Kustomizer) GetResource(kind, name string) (res KustomizerResource, ok bool) {
	res, ok = k.resourceMap[ResourceKey{Name: name, Kind: kind}]
	return
}
