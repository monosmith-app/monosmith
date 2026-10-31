package sbom

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Request asks for the SBOM of one released project.
type Request struct {
	Version string
	// Product names the file: <product>_<version>_sbom.cdx.json; the project's
	// name when empty.
	Product string
	// OutDir is the workspace-relative directory the file is written to.
	OutDir       string
	SmithVersion string
}

// Write asks the project's plugin for its components, writes the CycloneDX
// SBOM and returns its workspace-relative path. It returns "" when the plugin
// cannot list components, so the project gets no SBOM.
func Write(s *session.Session, p *projectgraph.Project, r Request) (string, error) {
	var listed protocol.ComponentsResult
	ok, err := s.Plugins.Call(p.Plugin, protocol.MethodComponents, protocol.ComponentsParams{Project: p.Project}, &listed, 2*time.Minute)
	if err != nil {
		return "", fmt.Errorf("sbom: %w", err)
	}
	if !ok {
		return "", nil
	}
	return writeListed(s.Root, p.Name, listed, r)
}

// writeListed renders and writes the SBOM of a project whose components the
// plugin listed.
func writeListed(root, name string, listed protocol.ComponentsResult, r Request) (string, error) {
	comps := make([]Component, len(listed.Components))
	for i, c := range listed.Components {
		comps[i] = Component{Name: c.Name, Version: c.Version, PURL: c.PURL}
	}
	sub := Subject{Name: name, Version: r.Version}
	if listed.PURL != "" {
		sub.PURL = listed.PURL + "@" + r.Version
	}
	body, err := RenderCycloneDX(sub, comps, time.Now(), r.SmithVersion)
	if err != nil {
		return "", err
	}
	product := r.Product
	if product == "" {
		product = name
	}
	rel := path.Join(r.OutDir, product+"_"+r.Version+"_sbom.cdx.json")
	return rel, writeFile(root, rel, append(body, '\n'))
}

func writeFile(root, rel string, body []byte) error {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, body, 0o644)
}
