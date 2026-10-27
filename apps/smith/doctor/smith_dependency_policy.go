package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const smithPackage = "monosmith"

var exactVersion = regexp.MustCompile(`^\d+\.\d+\.\d+([-+][0-9A-Za-z.+-]+)?$`)

// smithDependencyFinding checks that the smith the pipelines install is the
// smith the workspace's package.json depends on. The pipeline pins the smith
// that generated it unless [ci] smith_version says otherwise, so a newer local
// smith silently changes CI while the dependency stays behind. Only a Node
// dependency is checked; nothing is reported without one, for "latest", or
// when the versions agree. The version compared is the locked one
// (package-lock.json), else the declared one when it is exact.
func smithDependencyFinding(root, pipelineVersion string) []protocol.Finding {
	if pipelineVersion == "" || pipelineVersion == "latest" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil
	}
	section, declared := "", ""
	for _, s := range []string{"devDependencies", "dependencies"} {
		if deps := doc.Object(s, false); deps != nil {
			if v, ok := deps.String(smithPackage); ok {
				section, declared = s, v
				break
			}
		}
	}
	if section == "" {
		return nil
	}
	have := lockedVersion(root)
	if have == "" && exactVersion.MatchString(declared) {
		have = declared
	}
	if have == "" || have == pipelineVersion {
		return nil
	}
	doc.Object(section, true).SetString(smithPackage, pipelineVersion)
	return []protocol.Finding{{
		Check:   "the smith the pipelines install is the smith in your dependencies",
		Warning: true,
		Detail:  "the pipelines install " + smithPackage + " " + pipelineVersion + ", but package.json resolves " + have + ", so CI and local runs use different smiths",
		Remedy:  "`smith doctor --fix` sets the " + smithPackage + " dependency to " + pipelineVersion + " (then run your install to refresh the lockfile), or set [ci] smith_version = \"" + have + "\" and regenerate the pipelines",
		Fix: &protocol.Fix{
			Description: "set the " + smithPackage + " dependency to " + pipelineVersion,
			Files:       []protocol.FileEdit{{Path: "package.json", Content: string(doc.Bytes())}},
		},
	}}
}

// lockedVersion is the monosmith version package-lock.json installs, or "".
func lockedVersion(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, "package-lock.json"))
	if err != nil {
		return ""
	}
	var lock struct {
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if json.Unmarshal(raw, &lock) != nil {
		return ""
	}
	return lock.Packages["node_modules/"+smithPackage].Version
}
