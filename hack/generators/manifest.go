package generators

import (
	"bytes"
	"encoding/base64"
	"os"
	"path"
	"text/template"

	"ctx.sh/seaway/pkg/util/kustomize"
)

// InstallGenerator is used to generate the embedded installation manifests.
type InstallGenerator struct {
	OutputDir string
	ConfigDir string
}

// nolint:gochecknoglobals
var tmpl = `// Code generated by generator.go; DO NOT EDIT.
package install

import (
	"encoding/base64"
)

// Generated YAML for the CRD installation.
var crdYaml = {{ .CrdYAML }}

// Generated YAML for the controller installation.
var controllerYaml = {{ .ControllerYAML }}

// Generated YAML for a simple localstack installation.
var localstackYaml = {{ .LocalstackYAML }}

// Generated YAML for cert-manager installation.
var certManagerYaml = {{ .CertManagerYAML }}

// Generated YAML for registry installation.
var registryYaml = {{ .RegistryYAML }}

func decodeYAML(enc string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(enc)
}

func GetCrdBytes() ([]byte, error) {
	return decodeYAML(crdYaml)
}

func GetControllerBytes() ([]byte, error) {
	return decodeYAML(controllerYaml)
}

func GetLocalstackBytes() ([]byte, error) {
	return decodeYAML(localstackYaml)
}

func GetCertManagerBytes() ([]byte, error) {
	return decodeYAML(certManagerYaml)
}

func GetRegistryBytes() ([]byte, error) {
	return decodeYAML(registryYaml)
}
`

// Generate creates the file containing the install manifests and helper functions. The
// variables are base64 encoded as it is not uncommon to have values in yaml that need
// escaped.  Though I'm not a huge fan of this approach because it obscures what is being
// installed, but it's the most straightforward approach that covers all bases.  There will
// be a dry run available for users to dump the raw yaml without installing.
func (g *InstallGenerator) Generate() error {
	crdDocs, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: path.Join(g.ConfigDir, "crd"),
	})
	assertNoError(err)

	controllerDocs, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: path.Join(g.ConfigDir, "base"),
	})
	assertNoError(err)

	certManagerDocs, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: path.Join(g.ConfigDir, "cert-manager"),
	})
	assertNoError(err)

	localstackDocs, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: path.Join(g.ConfigDir, "localstack"),
	})
	assertNoError(err)

	registryDocs, err := kustomize.NewKustomizer(&kustomize.KustomizerOptions{
		BaseDir: path.Join(g.ConfigDir, "registry"),
	})
	assertNoError(err)

	t := template.Must(template.New("generators").Parse(tmpl))

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]string{
		"CrdYAML":         "`\n" + encode(crdDocs.Raw()) + "`",
		"ControllerYAML":  "`\n" + encode(controllerDocs.Raw()) + "`",
		"LocalstackYAML":  "`\n" + encode(localstackDocs.Raw()) + "`",
		"CertManagerYAML": "`\n" + encode(certManagerDocs.Raw()) + "`",
		"RegistryYAML":    "`\n" + encode(registryDocs.Raw()) + "`",
	})
	assertNoError(err)

	file := path.Join(g.OutputDir, "manifests.go")
	err = os.WriteFile(file, buf.Bytes(), 0644) // nolint:gosec
	assertNoError(err)

	return nil
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
