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

package v1beta1

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (m *Manifest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ManifestDefaulted Manifest
	var defaults = ManifestDefaulted{
		Version:     "v0.0.1",
		Description: "Application managed by Seaway",
	}

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	if out.Name == "" {
		return errors.New("name is a required field")
	}

	tmpl := Manifest(out)
	*m = tmpl
	return nil
}

// Load reads the manifest file and unmarshals it into the Manifest struct.
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

// GetEnvironment returns the development environment identified by name.
func (m *Manifest) GetEnvironment(name string) (ManifestEnvironmentSpec, error) {
	for _, e := range m.Environments {
		if e.Name == name {
			return e, nil
		}
	}

	return ManifestEnvironmentSpec{}, errors.New("environment not found")
}
