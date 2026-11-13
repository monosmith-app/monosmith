package tsconfigchain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// maxDepth stops a chain that loops.
const maxDepth = 16

// ReachesBase says whether a tsconfig file reaches the workspace's
// tsconfig.base.json through its "extends" chain. file is a path under root.
// A chain follows relative paths (a string or an array) and stops at a package
// name such as "astro/tsconfigs/strict", which cannot reach the base. A file
// that is not JSON gives false and the reason as an error; a missing file in
// the middle of a chain gives false.
func ReachesBase(root, file string) (bool, error) {
	base := filepath.Clean(filepath.Join(root, Base))
	return reaches(base, filepath.Clean(file), 0)
}

func reaches(base, file string, depth int) (bool, error) {
	if filepath.Clean(file) == base {
		return true, nil
	}
	if depth > maxDepth {
		return false, nil
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		if depth > 0 && os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	var doc struct {
		Extends json.RawMessage `json:"extends"`
	}
	if err := json.Unmarshal(jsondoc.StripComments(raw), &doc); err != nil {
		return false, err
	}
	for _, target := range extendsOf(doc.Extends) {
		if !strings.HasPrefix(target, ".") {
			continue
		}
		if !strings.HasSuffix(target, ".json") {
			target += ".json"
		}
		ok, err := reaches(base, filepath.Join(filepath.Dir(file), filepath.FromSlash(target)), depth+1)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// extendsOf reads the "extends" member: a string, or an array of strings.
func extendsOf(raw json.RawMessage) []string {
	var one string
	if json.Unmarshal(raw, &one) == nil && one != "" {
		return []string{one}
	}
	var many []string
	_ = json.Unmarshal(raw, &many)
	return many
}
