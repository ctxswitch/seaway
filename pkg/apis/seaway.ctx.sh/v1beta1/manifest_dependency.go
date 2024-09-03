package v1beta1

import "fmt"

// UnmarshalYAML implements the yaml.Unmarshaler interface.  This is used exclusively
// for the manifest loading process in the client.
func (me *ManifestDependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ManifestDependencyDefaulted ManifestDependency
	var defaults = ManifestDependencyDefaulted{
		Type: "kustomize",
	}

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	if out.Name == "" {
		return fmt.Errorf("dependency name is required")
	}

	if out.Path == "" {
		return fmt.Errorf("dependency path is required")
	}

	tmpl := ManifestDependency(out)
	*me = tmpl
	return nil
}
