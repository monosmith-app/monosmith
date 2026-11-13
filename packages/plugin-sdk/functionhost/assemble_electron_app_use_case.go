package functionhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// forgeConfig is the Electron Forge configuration of the app smith makes: it
// packs into the work folder's out/.
const forgeConfig = "module.exports = { outDir: '../out', packagerConfig: { asar: true } }\n"

// AssembleElectronApp writes the work folder that Electron runs and Forge
// packs: the manifest with the installed Electron version and the Forge
// configuration, and the page of the renderer. The compiled code is already
// in the folder.
//
// Parameters: projectDir is the project folder; appDir the work folder
// (dist/electron/<project dir>/app).
// Returns an error when Electron is not installed or a file cannot be written.
func AssembleElectronApp(projectDir, appDir string) error {
	raw, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
	if err != nil {
		return err
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return fmt.Errorf("package.json: %w", err)
	}
	version, err := installedVersion(projectDir, "electron")
	if err != nil {
		return err
	}
	manifest["devDependencies"] = map[string]any{"electron": version}
	manifest["config"] = map[string]any{"forge": "./forge.config.cjs"}
	renderer := filepath.Join(appDir, "dist", "renderer")
	if err := os.MkdirAll(renderer, 0o755); err != nil {
		return err
	}
	text, err := marshalManifest(manifest)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(appDir, "package.json"), text, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(appDir, "forge.config.cjs"), []byte(forgeConfig), 0o644); err != nil {
		return err
	}
	return copyFile(filepath.Join(projectDir, "src", "renderer", "index.html"), filepath.Join(renderer, "index.html"))
}

// installedVersion is the version of a package that npm installed, found in
// the node_modules of the folder or of a folder above it.
func installedVersion(from, name string) (string, error) {
	for dir := from; ; dir = filepath.Dir(dir) {
		raw, err := os.ReadFile(filepath.Join(dir, "node_modules", filepath.FromSlash(name), "package.json"))
		if err == nil {
			var m struct {
				Version string `json:"version"`
			}
			if err := json.Unmarshal(raw, &m); err != nil {
				return "", err
			}
			return m.Version, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("%s is not installed: run smith install first", name)
		}
	}
}
