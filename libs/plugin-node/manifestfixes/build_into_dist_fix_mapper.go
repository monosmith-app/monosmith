package manifestfixes

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// BuildIntoDistDescription names the fix on every package it applies to, so
// the host folds them into one.
const BuildIntoDistDescription = "build into the workspace's dist folder, with smith passing the output folder"

var (
	outDirWithComma = regexp.MustCompile(`(?s)[ \t]*"outDir"\s*:\s*"[^"]*"\s*,[ \t]*\r?\n?`)
	outDirLast      = regexp.MustCompile(`(?s),\s*"outDir"\s*:\s*"[^"]*"`)
	outDirValue     = regexp.MustCompile(`("outDir"\s*:\s*)"[^"]*"`)
	outDirScript    = regexp.MustCompile(`--outDir([= ]+)"?([^"\s&;|]+)"?`)
)

// BuildIntoDistFix moves a project's build output to the workspace's
// dist/<project dir>. configName is the tsconfig that sets oldOutDir (""
// when the build script's own --outDir flag does). When the build script runs
// tsc, the folder moves out of the tsconfig into a --outDir flag, so the
// tsconfig names no output folder; otherwise the tsconfig's outDir is
// pointed at dist. The package.json entries that read the old folder (main,
// types, module and the start script) follow it. It returns nil when there is
// nothing to edit.
func BuildIntoDistFix(workspaceRoot string, p protocol.Project, configName, oldOutDir string) *protocol.Fix {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	newDir := protocol.BuildDirFromProject(p.Root)
	manifest := path.Join(p.Root, "package.json")
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil
	}
	scripts := doc.Object("scripts", false)
	build := ""
	if scripts != nil {
		build, _ = scripts.String("build")
	}
	fix := &protocol.Fix{Description: BuildIntoDistDescription, AppliesTo: []string{p.Name}}

	switch {
	case configName == "":
		if scripts == nil || !outDirScript.MatchString(build) {
			return nil
		}
		scripts.SetString("build", outDirScript.ReplaceAllString(build, "--outDir${1}"+newDir))
	case scripts != nil && strings.Contains(build, "tsc"):
		config, err := os.ReadFile(filepath.Join(dir, configName))
		if err != nil {
			return nil
		}
		edited := dropOutDir(string(config))
		if edited == string(config) {
			return nil
		}
		fix.Files = append(fix.Files, protocol.FileEdit{Path: path.Join(p.Root, configName), Content: edited})
		scripts.SetString("build", addOutDirFlag(build, newDir))
	default:
		config, err := os.ReadFile(filepath.Join(dir, configName))
		if err != nil {
			return nil
		}
		edited := outDirValue.ReplaceAllString(string(config), `${1}"`+newDir+`"`)
		if edited == string(config) {
			return nil
		}
		fix.Files = append(fix.Files, protocol.FileEdit{Path: path.Join(p.Root, configName), Content: edited})
	}

	old := strings.TrimPrefix(path.Clean(strings.ReplaceAll(oldOutDir, "\\", "/")), "./")
	for _, key := range []string{"main", "types", "typings", "module"} {
		if v, ok := doc.String(key); ok {
			doc.SetString(key, followOutDir(v, old, newDir))
		}
	}
	if scripts != nil {
		if start, ok := scripts.String("start"); ok {
			scripts.SetString("start", followOutDir(start, old, newDir))
		}
	}
	fix.Files = append(fix.Files, protocol.FileEdit{Path: manifest, Content: render(doc, raw)})
	return fix
}

// dropOutDir removes the outDir property from the text of a tsconfig.
func dropOutDir(config string) string {
	if outDirWithComma.MatchString(config) {
		return outDirWithComma.ReplaceAllString(config, "")
	}
	return outDirLast.ReplaceAllString(config, "")
}

// addOutDirFlag adds --outDir to the first tsc command of a script.
func addOutDirFlag(script, dir string) string {
	parts := strings.Split(script, " && ")
	for i, part := range parts {
		if strings.Contains(part, "tsc") {
			parts[i] = part + " --outDir " + dir
			break
		}
	}
	return strings.Join(parts, " && ")
}

// followOutDir rewrites a value that starts with the old output folder
// ("./dist/main.js", "node dist/main.js") to the new one.
func followOutDir(value, old, newDir string) string {
	if old == "" || old == "." {
		return value
	}
	for _, prefix := range []string{"./" + old + "/", old + "/"} {
		if strings.HasPrefix(value, prefix) {
			return newDir + "/" + strings.TrimPrefix(value, prefix)
		}
		if i := strings.Index(value, " "+prefix); i >= 0 {
			return value[:i+1] + newDir + "/" + value[i+1+len(prefix):]
		}
	}
	return value
}

// comments strips // and /* */ comments, which tsconfig allows.
var comments = regexp.MustCompile(`(?s)/\*.*?\*/|(?m)^[ \t]*//[^\n]*`)

// LocateOutDir finds where the build of the project in dir writes: the
// --outDir flag of its build script (config is "" then), or the outDir of
// tsconfig.build.json, then tsconfig.json (config names the file). ok is
// false when the project sets no output folder.
func LocateOutDir(dir string) (config, outDir string, ok bool) {
	if raw, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var manifest struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(raw, &manifest) == nil {
			if m := outDirScript.FindStringSubmatch(manifest.Scripts["build"]); m != nil {
				return "", m[2], true
			}
		}
	}
	for _, name := range []string{"tsconfig.build.json", "tsconfig.json"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var cfg struct {
			CompilerOptions struct {
				OutDir string `json:"outDir"`
			} `json:"compilerOptions"`
		}
		if json.Unmarshal(comments.ReplaceAll(raw, nil), &cfg) != nil || cfg.CompilerOptions.OutDir == "" {
			continue
		}
		return name, cfg.CompilerOptions.OutDir, true
	}
	return "", "", false
}
