package core

import (
	"log"
	"os"
	"regexp"
	"strings"

	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

type Seaway struct {
	Endpoint string   `yaml:"endpoint"`
	UseSSL   bool     `yaml:"useSSL"`
	Image    string   `yaml:"image"`
	Bucket   string   `yaml:"bucket"`
	Prefix   string   `yaml:"prefix"`
	Include  []string `yaml:"include"`
	Exclude  []string `yaml:"exclude"`
}

func (s *Seaway) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type SeawayDefaulted Seaway
	var defaults = SeawayDefaulted{
		Endpoint: "localhost:9000",
		UseSSL:   true,
		Bucket:   "seaway",
		Prefix:   "dev",
	}

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	tmpl := Seaway(out)
	*s = tmpl
	return nil
}

type Manifest struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Seaway      Seaway `yaml:"seaway"`
}

func NewManifest() *Manifest {
	return &Manifest{}
}

func (m *Manifest) Load(file string) error {
	manifestFile, err := os.ReadFile("manifest.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(manifestFile, m)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (m *Manifest) Includes() *regexp.Regexp {
	include := append(DefaultIncludes, m.Seaway.Include...)
	r := strings.Join(include, "|")
	return regexp.MustCompile(r)
}

func (m *Manifest) Excludes() *regexp.Regexp {
	exclude := append(DefaultExcludes, m.Seaway.Exclude...)
	r := strings.Join(exclude, "|")
	return regexp.MustCompile(r)
}
