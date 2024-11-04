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
	"regexp"
	"strings"
)

const (
	DefaultEndpoint = "http://localhost:8080"
)

var (
	DefaultExcludes = []string{
		"vendor/*",
		".venv/*",
		"node_modules/*",
		".git/*",
		".idea/*",
		".vscode/*",
		".terraform/*",
	}

	DefaultIncludes = []string{
		"manifest.yaml",
		"Dockerfile",
	}
)

// UnmarshalYAML implements the yaml.Unmarshaler interface.  This is used exclusively
// for the manifest loading process in the client.
func (me *ManifestEnvironmentSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ManifestEnvironmentSpecDefaulted ManifestEnvironmentSpec
	var defaults = ManifestEnvironmentSpecDefaulted{
		Namespace: "default",
		Endpoint:  DefaultEndpoint,
	}

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	tmpl := ManifestEnvironmentSpec(out)
	*me = tmpl
	return nil
}

// Includes returns a regular expression that matches the files that should be included
// in the build context.  It is used while when we walk the file system to build the
// archive that will be uploaded to the object storage.
func (me *ManifestEnvironmentSpec) Includes() *regexp.Regexp {
	include := append(DefaultIncludes, me.Build.Include...) //nolint:gocritic
	r := strings.Join(include, "|")
	return regexp.MustCompile(r)
}

// Excludes returns a regular expression that matches the files that should be excluded
// from the build context.  Like includes, it is used when we walk the file system to
// build the archive that will be uploaded to the object storage.  Currently, excludes
// are processed after includes so if there are files in included directories that match
// the exclude pattern they will be excluded.
func (me *ManifestEnvironmentSpec) Excludes() *regexp.Regexp {
	exclude := append(DefaultExcludes, me.Build.Exclude...) //nolint:gocritic
	r := strings.Join(exclude, "|")
	return regexp.MustCompile(r)
}
