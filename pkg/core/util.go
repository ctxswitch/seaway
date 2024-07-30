package core

func MetadataPath(manifest *Manifest) string {
	return "_metadata/" + manifest.Name + "/metadata.json"
}

func MetadataLockPath(manifest *Manifest) string {
	return "_metadata/" + manifest.Name + "/metadata.lock"
}
