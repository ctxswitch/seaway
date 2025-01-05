package util

import "strings"

func ArchiveKey(prefix, namespace, name string) string {
	nsn := namespace + "-" + name + ".tar.gz"
	return strings.Join([]string{prefix, nsn}, "/")
}
