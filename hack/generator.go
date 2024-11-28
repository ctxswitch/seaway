//go:generate go run generator.go ../config/seaway/base

package main

import (
	"os"

	"ctx.sh/seaway/hack/generators"
)

func main() {
	configDir := os.Args[1]
	outputDir := os.Args[2]
	gen := generators.InstallGenerator{
		ConfigDir: configDir,
		OutputDir: outputDir,
	}
	_ = gen.Generate()
}
