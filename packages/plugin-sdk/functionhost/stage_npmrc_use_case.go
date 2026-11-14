package functionhost

import (
	"os"
	"path/filepath"
)

// stageNpmrc gives the install in the stage the .npmrc that applies to the
// project. The stage is outside the workspace, so npm would not read the
// workspace's .npmrc there: a scope mapped to a private feed (`@acme:registry=...`)
// is lost, and its packages answer 404 from the public registry. The file is
// copied with its mode kept private, as it can name a token variable.
//
// The nearest .npmrc is used, from the project folder up to the workspace root
// (the folder with smith.toml or .git); the folder of the person's home is never
// taken, since npm reads that file by itself.
//
// Parameters: projectDir is the project folder; stage is the folder npm installs in.
// Returns: remove, which deletes the copy and must run before the stage is zipped
// (a token line must never reach the zip); an error when the copy fails.
func stageNpmrc(projectDir, stage string) (remove func(), err error) {
	source := nearestNpmrc(projectDir)
	if source == "" {
		return func() {}, nil
	}
	raw, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	target := filepath.Join(stage, ".npmrc")
	if err := os.WriteFile(target, raw, 0o600); err != nil {
		return nil, err
	}
	return func() { _ = os.Remove(target) }, nil
}

// nearestNpmrc is the first .npmrc from dir up to the workspace root; "" when none.
func nearestNpmrc(dir string) string {
	home, _ := os.UserHomeDir()
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		if dir != home {
			if info, err := os.Stat(filepath.Join(dir, ".npmrc")); err == nil && !info.IsDir() {
				return filepath.Join(dir, ".npmrc")
			}
		}
		if exists(filepath.Join(dir, "smith.toml")) || exists(filepath.Join(dir, ".git")) {
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
