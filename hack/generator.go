//go:generate go run generator.go ../config/seaway/base

package main

import (
	"fmt"
	"os"

	"ctx.sh/seaway/hack/generators"
)

func main() {
	fmt.Printf("Args: %v\n", os.Args)
	configDir := os.Args[1]
	outputDir := os.Args[2]
	gen := generators.InstallGenerator{
		ConfigDir: configDir,
		OutputDir: outputDir,
	}
	_ = gen.Generate()
}
