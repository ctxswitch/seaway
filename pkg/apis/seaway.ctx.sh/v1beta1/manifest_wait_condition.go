package v1beta1

import (
	"fmt"
	"time"
)

func (mw *ManifestWaitCondition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type ManifestWaitConditionDefaulted ManifestWaitCondition
	var defaults = ManifestWaitConditionDefaulted{
		Timeout: 5 * time.Minute,
		For:     "ready",
	}

	out := defaults
	if err := unmarshal(&out); err != nil {
		return err
	}

	if out.Kind == "" {
		return fmt.Errorf("kind is required")
	}

	if out.Name == "" {
		return fmt.Errorf("name is required")
	}

	tmpl := ManifestWaitCondition(out)
	*mw = tmpl
	return nil
}
