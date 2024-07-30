package core

var (
	DefaultExcludes = []string{
		"vendor/*",
		".venv/*",
		"node_modules/*",
		".git/*",
		".idea/*",
		".vscode/*",
		".terraform/*",
		"seaway.sum",
	}

	DefaultIncludes = []string{
		"manifest.yaml",
	}
)
