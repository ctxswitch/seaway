package v1beta1

import (
	"regexp"
	"strings"
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

func (me *ManifestEnvironmentSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ManifestEnvironmentSpecDefaulted ManifestEnvironmentSpec
	var defaults = ManifestEnvironmentSpecDefaulted{
		Namespace: "default",
	}

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	// When we use the manifest for syncing we only need the source information to
	// be defaulted.  Other fields such as the build.platform need to be defaulted
	// in the k8s environment that they will be built in to get the correct values.
	defaultEnvironmentSource(out.Source)

	tmpl := ManifestEnvironmentSpec(out)
	*me = tmpl
	return nil
}

// TODO: redo this..
func (me *ManifestEnvironmentSpec) Includes() *regexp.Regexp {
	include := append(DefaultIncludes, me.Build.Include...)
	r := strings.Join(include, "|")
	return regexp.MustCompile(r)
}

func (me *ManifestEnvironmentSpec) Excludes() *regexp.Regexp {
	exclude := append(DefaultExcludes, me.Build.Exclude...)
	r := strings.Join(exclude, "|")
	return regexp.MustCompile(r)
}
