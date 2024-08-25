package v1beta1

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func (m *Manifest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ManifestDefaulted Manifest
	var defaults = ManifestDefaulted{
		Version:     "v0.0.1",
		Description: "Application managed by Seaway",
	}

	// TODO: Validate name is present...

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	tmpl := Manifest(out)
	*m = tmpl
	return nil
}

func (m *Manifest) Load(file string) error {
	manifest, err := os.ReadFile("manifest.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(manifest, m)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (m *Manifest) GetEnvironment(name string) (ManifestEnvironmentSpec, error) {
	for _, e := range m.Environments {
		if e.Name == name {
			return e, nil
		}
	}

	return ManifestEnvironmentSpec{}, errors.New("environment not found")
}
