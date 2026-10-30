package declaredimports

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// installedModules reads the workspace venv's metadata: for each installed
// distribution (normalised name), the top-level modules it provides. A venv
// that is missing, or a distribution it lacks, simply has no entry.
func installedModules(workspaceRoot string) map[string][]string {
	out := map[string][]string{}
	var sites []string
	for _, pattern := range []string{"Lib/site-packages", "lib/python*/site-packages", "lib64/python*/site-packages"} {
		matches, _ := filepath.Glob(filepath.Join(workspaceRoot, ".venv", filepath.FromSlash(pattern)))
		sites = append(sites, matches...)
	}
	for _, site := range sites {
		infos, _ := filepath.Glob(filepath.Join(site, "*.dist-info"))
		for _, info := range infos {
			stem := strings.TrimSuffix(filepath.Base(info), ".dist-info")
			name, _, ok := strings.Cut(stem, "-")
			if !ok {
				continue
			}
			if modules := distributionModules(info); len(modules) > 0 {
				out[normalise(name)] = modules
			}
		}
	}
	return out
}

// distributionModules are the top-level modules of one .dist-info: its
// top_level.txt, or else the first path segments of its RECORD.
func distributionModules(info string) []string {
	if raw, err := os.ReadFile(filepath.Join(info, "top_level.txt")); err == nil {
		return strings.Fields(string(raw))
	}
	raw, err := os.ReadFile(filepath.Join(info, "RECORD"))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		file, _, _ := strings.Cut(line, ",")
		first, _, nested := strings.Cut(path.Clean(strings.ReplaceAll(file, "\\", "/")), "/")
		switch {
		case !nested && strings.HasSuffix(first, ".py"):
			first = strings.TrimSuffix(first, ".py")
		case !nested, strings.HasSuffix(first, ".dist-info"), strings.HasSuffix(first, ".data"), first == "..", first == "__pycache__":
			continue
		}
		if !seen[first] {
			seen[first] = true
			out = append(out, first)
		}
	}
	return out
}
