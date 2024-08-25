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

type KustomizerResource struct {
	Resource *unstructured.Unstructured
	GVK      *schema.GroupVersionKind
}

type KustomizerOptions struct {
	BaseDir string
}

type Kustomizer struct {
	docs *utilyaml.YAMLReader
}

func NewKustomizer(opts *KustomizerOptions) (*Kustomizer, error) {
	kustomizer := krusty.MakeKustomizer(&krusty.Options{
		Reorder:           "none",
		AddManagedbyLabel: false,
		LoadRestrictions:  types.LoadRestrictionsRootOnly,
		PluginConfig:      &types.PluginConfig{},
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
		docs: utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(yml))),
	}, nil
}

// TODO: Add the envsubst capability where we can substitute after reading the
// doc.
func (k *Kustomizer) Resources() ([]KustomizerResource, error) {
	resources := []KustomizerResource{}
	for {
		doc, err := k.docs.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		decoded := &unstructured.Unstructured{}

		_, gvk, err := decoder.Decode(doc, nil, decoded)
		if err != nil {
			return nil, err
		}

		resources = append(resources, KustomizerResource{
			Resource: decoded,
			GVK:      gvk,
		})
	}

	return resources, nil
}

func (k *Kustomizer) Next() (KustomizerResource, bool, error) {
	doc, err := k.docs.Read()
	if err != nil {
		if err.Error() == "EOF" {
			return KustomizerResource{}, true, nil
		}
		return KustomizerResource{}, true, err
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	decoded := &unstructured.Unstructured{}

	_, gvk, err := decoder.Decode(doc, nil, decoded)
	return KustomizerResource{
		Resource: decoded,
		GVK:      gvk,
	}, false, err
}
