package v1beta1

import "fmt"

func (c *SeawayConfigStorageSpec) GetArchiveKey(name, namespace string) string {
	if c.Prefix == "" {
		return fmt.Sprintf("%s-%s.tar.gz", name, namespace)
	}

	return fmt.Sprintf("%s/%s-%s.tar.gz", c.Prefix, name, namespace)
}
