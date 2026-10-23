package sbom

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// The CycloneDX 1.5 JSON schema and the two it refers to (testdata, published
// by the CycloneDX project under the Apache License 2.0): an SBOM smith writes
// must be one a consumer's validator accepts, not only one with the right keys.
func cyclonedxSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	for _, name := range []string{"bom-1.5.schema.json", "spdx.schema.json", "jsf-0.82.schema.json"} {
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		if err := c.AddResource("http://cyclonedx.org/schema/"+name, doc); err != nil {
			t.Fatal(err)
		}
	}
	schema, err := c.Compile("http://cyclonedx.org/schema/bom-1.5.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func validate(t *testing.T, schema *jsonschema.Schema, raw []byte) {
	t.Helper()
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("not JSON: %v\n%s", err, raw)
	}
	if err := schema.Validate(doc); err != nil {
		t.Errorf("not a valid CycloneDX 1.5 document: %v\n%s", err, raw)
	}
}

func TestRenderedSBOMIsValidCycloneDX(t *testing.T) {
	schema := cyclonedxSchema(t)
	at := time.Date(2026, 10, 23, 1, 0, 0, 0, time.UTC)
	subject := Subject{Name: "app", Version: "1.0.0", PURL: "pkg:golang/example.com/app@1.0.0"}
	cases := map[string][]Component{
		"no components": nil,
		"components": {
			{Name: "github.com/a/b", Version: "v1.2.3", PURL: "pkg:golang/github.com/a/b@v1.2.3"},
			{Name: "@scope/pkg", Version: "2.0.0", PURL: "pkg:npm/%40scope/pkg@2.0.0"},
			{Name: "no-version", PURL: "pkg:npm/no-version"},
		},
	}
	for name, comps := range cases {
		raw, err := RenderCycloneDX(subject, comps, at, "1.2.3")
		if err != nil {
			t.Fatal(err)
		}
		t.Run(name, func(t *testing.T) { validate(t, schema, raw) })
	}
}

func TestWrittenSBOMsAreValidCycloneDX(t *testing.T) {
	schema := cyclonedxSchema(t)
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "app"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "app", "go.mod"), []byte("module example.com/app\n\ngo 1.22\n\nrequire github.com/a/b v1.2.3\n"), 0o644)
	rel, err := Write(Request{Root: root, ProjectRoot: "app", Language: "go", Name: "app", Version: "1.0.0", OutDir: "dist/release/app", SmithVersion: "x"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	validate(t, schema, raw)
}

// A schema that accepted anything would make the tests above pass for nothing.
func TestTheSchemaRejectsWhatIsNotCycloneDX(t *testing.T) {
	schema := cyclonedxSchema(t)
	for name, raw := range map[string]string{
		"no spec version": `{"bomFormat":"CycloneDX"}`,
		"wrong format":    `{"bomFormat":"SPDX","specVersion":"1.5","version":1}`,
		"component type":  `{"bomFormat":"CycloneDX","specVersion":"1.5","version":1,"components":[{"type":"gadget","name":"x"}]}`,
	} {
		doc, _ := jsonschema.UnmarshalJSON(bytes.NewReader([]byte(raw)))
		if schema.Validate(doc) == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
