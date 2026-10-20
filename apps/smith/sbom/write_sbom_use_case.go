package sbom

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Request asks for the SBOM of one released project.
type Request struct {
	// Root is the workspace directory; ProjectRoot is the project's, relative.
	Root, ProjectRoot string
	Language          string
	Name, Version     string
	// Product names the file: <product>_<version>_sbom.cdx.json.
	Product string
	// OutDir is the workspace-relative directory the file is written to.
	OutDir       string
	SmithVersion string
}

// Write writes the project's CycloneDX SBOM and returns its workspace-relative
// path, or "" when the project's ecosystem is not supported yet.
func Write(r Request) (string, error) {
	sub := Subject{Name: r.Name, Version: r.Version}
	var comps []Component
	var err error
	switch r.Language {
	case "go":
		var gomod []byte
		if gomod, err = os.ReadFile(filepath.Join(r.Root, filepath.FromSlash(r.ProjectRoot), "go.mod")); err != nil {
			return "", fmt.Errorf("sbom: %w", err)
		}
		comps = GoComponents(string(gomod))
		if m := GoModule(string(gomod)); m != "" {
			sub.PURL = "pkg:golang/" + m + "@" + r.Version
		}
	case "node":
		if comps, err = nodeComponents(r); err != nil {
			return "", fmt.Errorf("sbom: %w", err)
		}
	default:
		return "", nil
	}
	body, err := RenderCycloneDX(sub, comps, time.Now(), r.SmithVersion)
	if err != nil {
		return "", err
	}
	product := r.Product
	if product == "" {
		product = r.Name
	}
	rel := path.Join(r.OutDir, product+"_"+r.Version+"_sbom.cdx.json")
	abs := filepath.Join(r.Root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	return rel, os.WriteFile(abs, append(body, '\n'), 0o644)
}

// nodeComponents reads the nearest package-lock.json, at the project or any
// directory above it up to the workspace (npm workspaces keep one at the
// root), and falls back to the project's package.json.
func nodeComponents(r Request) ([]Component, error) {
	dir := filepath.Join(r.Root, filepath.FromSlash(r.ProjectRoot))
	for d := dir; ; d = filepath.Dir(d) {
		if lock, err := os.ReadFile(filepath.Join(d, "package-lock.json")); err == nil {
			return NPMComponentsFromLock(lock)
		}
		if d == filepath.Clean(r.Root) || filepath.Dir(d) == d {
			break
		}
	}
	manifest, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return NPMComponentsFromManifest(manifest)
}
