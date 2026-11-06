package legacyjesthelpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const oldResolver = "// Written by smith convert: what @nx/jest/plugins/resolver did when Jest cannot\n// find a module, without Nx.\nmodule.exports = function resolver () {}\n"
const oldTransform = "// Written by smith convert: what @nx/react/plugins/jest did, without Nx.\nmodule.exports = {}\n"
const oldPreset = "const nxPreset = " + legacyPresetValues + "\n\nmodule.exports = {\n  ...nxPreset,\n  setupFilesAfterEnv: [require.resolve('./jest.setup.cjs')],\n}\n"
const reactConfig = "module.exports = {\n  preset: '../../jest.preset.js',\n  transform: { '^(?!.*\\\\.(js|jsx|ts|tsx|css|json)$)': '<rootDir>/../../jest.asset-transform.cjs' },\n}\n"

func workspace(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, text := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func edited(fix *protocol.Fix, path string) (protocol.FileEdit, bool) {
	for _, f := range fix.Files {
		if f.Path == path {
			return f, true
		}
	}
	return protocol.FileEdit{}, false
}

func TestFindCleanupReplacesTheHelpersWithThePackage(t *testing.T) {
	root := workspace(t, map[string]string{
		"jest.resolver.cjs":        oldResolver,
		"jest.asset-transform.cjs": oldTransform,
		"jest.preset.js":           oldPreset,
		"apps/ui/jest.config.cts":  reactConfig,
		"package.json":             `{"name": "acme", "devDependencies": {"jest": "^30.0.0"}}` + "\n",
	})

	got := FindCleanup(root, []string{"apps/ui"}, "2.0.0-beta.2")

	if got.Fix == nil {
		t.Fatalf("no fix: %s", got.Manual)
	}
	if strings.Join(got.Files, ",") != "jest.resolver.cjs,jest.asset-transform.cjs" {
		t.Errorf("files %v", got.Files)
	}
	preset, _ := edited(got.Fix, "jest.preset.js")
	if !strings.Contains(preset.Content, "const nxPreset = require('@monosmith/jest-preset')") || !strings.Contains(preset.Content, "setupFilesAfterEnv: [require.resolve('./jest.setup.cjs')]") || strings.Contains(preset.Content, "testMatch") {
		t.Errorf("preset:\n%s", preset.Content)
	}
	config, _ := edited(got.Fix, "apps/ui/jest.config.cts")
	if !strings.Contains(config.Content, "'@monosmith/jest-preset/asset-transform'") || strings.Contains(config.Content, "jest.asset-transform.cjs") {
		t.Errorf("config:\n%s", config.Content)
	}
	manifest, _ := edited(got.Fix, "package.json")
	if !strings.Contains(manifest.Content, `"@monosmith/jest-preset": "^2.0.0-beta.1"`) {
		t.Errorf("manifest:\n%s", manifest.Content)
	}
	for _, gone := range []string{"jest.resolver.cjs", "jest.asset-transform.cjs"} {
		if f, ok := edited(got.Fix, gone); !ok || !f.Delete {
			t.Errorf("%s is not deleted: %+v", gone, f)
		}
	}
	if got.Fix.Safe {
		t.Error("the fix edits the user's Jest configs, so it is asked about")
	}
	if len(got.Fix.Steps) != 1 || got.Fix.Steps[0].Argv[0] != "npm" {
		t.Errorf("steps %v", got.Fix.Steps)
	}
}

func TestFindCleanupKeepsCRLFAndAnExistingDependency(t *testing.T) {
	root := workspace(t, map[string]string{
		"jest.resolver.cjs": oldResolver,
		"jest.preset.js":    strings.ReplaceAll("module.exports = "+legacyPresetValues+"\n", "\n", "\r\n"),
		"package.json":      `{"devDependencies": {"@monosmith/jest-preset": "^0.1.2"}}` + "\n",
	})

	got := FindCleanup(root, nil, "1.0.0")

	if got.Fix == nil {
		t.Fatalf("no fix: %s", got.Manual)
	}
	preset, _ := edited(got.Fix, "jest.preset.js")
	if preset.Content != "module.exports = require('@monosmith/jest-preset')\r\n" {
		t.Errorf("preset %q", preset.Content)
	}
	if _, ok := edited(got.Fix, "package.json"); ok {
		t.Error("the package is already a dependency: package.json stays as it is")
	}
}

func TestFindCleanupLeavesWhatItDoesNotRecognise(t *testing.T) {
	edited := strings.Replace(oldPreset, "testEnvironment:          'jsdom'", "testEnvironment:          'node'", 1)
	root := workspace(t, map[string]string{
		"jest.resolver.cjs": oldResolver,
		"jest.preset.js":    edited,
		"package.json":      "{}\n",
	})

	got := FindCleanup(root, nil, "1.0.0")

	if got.Fix != nil {
		t.Fatalf("a preset the user edited must not be rewritten: %+v", got.Fix)
	}
	if !strings.Contains(got.Manual, "jest.preset.js") || !strings.Contains(got.Manual, "@monosmith/jest-preset") {
		t.Errorf("manual: %s", got.Manual)
	}
}

func TestFindCleanupIgnoresFilesTheUserWrote(t *testing.T) {
	root := workspace(t, map[string]string{
		"jest.resolver.cjs": "module.exports = require('./my-resolver')\n",
		"package.json":      "{}\n",
	})

	if got := FindCleanup(root, nil, "1.0.0"); len(got.Files) != 0 || got.Fix != nil {
		t.Fatalf("a resolver smith did not write is not smith's to remove: %+v", got)
	}
	if got := FindCleanup(t.TempDir(), nil, "1.0.0"); len(got.Files) != 0 || got.Fix != nil {
		t.Fatalf("a workspace without the helpers has nothing to clean: %+v", got)
	}
}
