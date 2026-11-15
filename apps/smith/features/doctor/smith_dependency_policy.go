package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctorchecks"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

const smithPackage = "monosmith"

// launcherRange is the dependency on the npm launcher (ADR 0101).
const launcherRange = "^3"

var exactVersion = regexp.MustCompile(`^\d+\.\d+\.\d+([-+][0-9A-Za-z.+-]+)?$`)

// smithDependencyFinding checks a workspace's npm dependency on monosmith.
// From 3.0 the package is the launcher, which runs the smith smith.toml pins
// ([workspace] smith); below 3 it is smith itself, which no longer gets new
// versions on npm. An old dependency is moved to the launcher, and when
// smith.toml pins nothing, the version it resolved becomes the pin, so the
// same smith keeps running. The version compared is the locked one
// (package-lock.json), else the declared one when it is exact.
//
// Parameters: root is the workspace root; pin is [workspace] smith.
// Returns: one warning with its fix, or nothing.
func smithDependencyFinding(root, pin string) []protocol.Finding {
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
	if major, err := strconv.Atoi(strings.SplitN(have, ".", 2)[0]); have == "" || err != nil || major >= 3 {
		return nil
	}
	doc.Object(section, true).SetString(smithPackage, launcherRange)
	files := []protocol.FileEdit{{Path: "package.json", Content: string(doc.Bytes())}}
	remedy := "`smith doctor --fix` sets the " + smithPackage + " dependency to " + launcherRange + " and runs npm install (it asks first)"
	if pin == "" {
		if text, err := os.ReadFile(filepath.Join(root, workspace.FileName)); err == nil {
			files = append(files, protocol.FileEdit{Path: workspace.FileName, Content: tomltext.SetString(string(text), "workspace", "smith", have)})
			remedy += ", and pins [workspace] smith = \"" + have + "\" so the same smith keeps running"
		}
	}
	return []protocol.Finding{{
		Check:   "the monosmith dependency is the launcher",
		Warning: true,
		Area:    doctorchecks.AreaDependencies,
		Problem: "package.json resolves monosmith " + have + ", which is smith itself; from 3.0 the package is the launcher that runs the smith smith.toml pins",
		Detail:  "smith no longer publishes its versions to npm, so this dependency stays on " + have + " for good (ADR 0101)",
		Remedy:  remedy,
		Fix: &protocol.Fix{
			Description: "depend on the monosmith launcher (" + launcherRange + ")",
			Files:       files,
			Steps: []protocol.FixStep{
				{Description: "run npm install to refresh package-lock.json", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}},
			},
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
