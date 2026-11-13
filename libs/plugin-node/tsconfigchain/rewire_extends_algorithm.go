package tsconfigchain

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// presetPrefix starts the names of the presets smith's own base extends.
const presetPrefix = "@monosmith/tsconfig/"

// Rewire returns the text of the tsconfig.json of the project in dir with an
// "extends" that reaches the workspace's tsconfig.base.json. A file without
// "extends", or one that extends only @monosmith/tsconfig presets (which the
// base extends too), gets the base alone. A file that extends something else
// (Astro's preset, a file of its own) gets the base first and its own entries
// after, so what that preset sets still wins over the base. Comments and the
// rest of the file stay as they are. It returns an error when src is not a
// JSON object.
func Rewire(src []byte, dir string) ([]byte, error) {
	var doc struct {
		Extends json.RawMessage `json:"extends"`
	}
	if err := json.Unmarshal(jsondoc.StripComments(src), &doc); err != nil {
		return nil, err
	}
	base := Extends(dir, Base)
	var value any = base
	own := extendsOf(doc.Extends)
	var keep []string
	for _, target := range own {
		if !strings.HasPrefix(target, presetPrefix) {
			keep = append(keep, target)
		}
	}
	if len(keep) > 0 {
		value = append([]string{base}, keep...)
	}
	return jsondoc.SetKeysInPlace(src, map[string]any{"extends": value})
}
