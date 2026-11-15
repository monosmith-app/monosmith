package discovery

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	scriptOutDir  = regexp.MustCompile(`--out(?:Dir|dir|-dir)([= ]+)"?([^"\s&;|]+)"?`)
	scriptOutFile = regexp.MustCompile(`--out(?:file|File)([= ]+)"?([^"\s&;|]+)"?`)
	scriptProject = regexp.MustCompile(`(?:^|\s)(?:-p|--project)[= ]+"?([^"\s&;|]+\.json)"?`)
)

// comments strips // and /* */ comments, which tsconfig allows.
var comments = regexp.MustCompile(`(?s)/\*.*?\*/|(?m)^[ \t]*//[^\n]*`)

// LocateOutDir finds where the build of the project in dir writes: the
// --outDir flag of its build script (config is "" then), or the outDir of
// tsconfig.build.json, then tsconfig.json (config names the file). ok is
// false when the project sets no output folder.
func LocateOutDir(dir string) (config, outDir string, ok bool) {
	candidates := []string{"tsconfig.build.json", "tsconfig.json"}
	if raw, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var manifest struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(raw, &manifest) == nil {
			build := manifest.Scripts["build"]
			if m := scriptOutDir.FindStringSubmatch(build); m != nil {
				return "", m[2], true
			}
			if m := scriptOutFile.FindStringSubmatch(build); m != nil {
				return "", path.Dir(strings.ReplaceAll(m[2], "\\", "/")), true
			}
			if m := scriptProject.FindStringSubmatch(build); m != nil {
				// tsc -p tsconfig.lib.json: the file the script names comes first.
				candidates = append([]string{m[1]}, candidates...)
			}
		}
	}
	for _, name := range candidates {
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
