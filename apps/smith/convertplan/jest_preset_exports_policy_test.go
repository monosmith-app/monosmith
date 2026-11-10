package convertplan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The specifiers convert writes into a project's Jest config must resolve in
// the package they name. Without an exports map, node cannot find
// "@monosmith/jest-preset/asset-transform" for a file called asset-transform.cjs.
func TestSpecifiersConvertWritesResolveInTheJestPresetPackage(t *testing.T) {
	root := filepath.Join("..", "..", "..", "packages", "jest-preset")
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Skipf("the jest-preset package is not next to this checkout: %v", err)
	}
	var pkg struct {
		Name    string            `json:"name"`
		Exports map[string]string `json:"exports"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Name != JestPresetPackage {
		t.Fatalf("the package is %q, convert writes %q", pkg.Name, JestPresetPackage)
	}
	written := []string{JestPresetPackage, AssetTransformModule, JestPresetPackage + "/react"}
	for _, specifier := range written {
		subpath := "." + strings.TrimPrefix(specifier, JestPresetPackage)
		target, ok := pkg.Exports[subpath]
		if !ok {
			t.Errorf("convert writes %q but the package does not export %q", specifier, subpath)
			continue
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(target))); err != nil {
			t.Errorf("%q exports %q, which does not exist: %v", specifier, target, err)
		}
	}
}
