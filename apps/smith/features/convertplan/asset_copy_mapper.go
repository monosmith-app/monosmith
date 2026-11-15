package convertplan

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/nxworkspace"
)

// assetCopyScript copies an Nx build's assets after the build, as Nx's
// copyAssets did: each entry is [input, glob, output, whole], paths relative
// to the project. A plain path entry (whole) copies a directory with all it
// holds; a glob entry copies the files it matches. It runs as `node -e`, so
// it works in every shell npm uses (fs.globSync needs Node 22 or later).
const assetCopyScript = `const fs=require('node:fs'),p=require('node:path');` +
	`for(const[i,g,o,w]of%s)for(const f of fs.globSync(g,{cwd:i})){const s=p.join(i,f),d=p.join(o,f);` +
	`if(fs.statSync(s).isDirectory()){if(w)fs.cpSync(s,d,{recursive:true})}` +
	`else{fs.mkdirSync(p.dirname(d),{recursive:true});fs.copyFileSync(s,d)}}`

// AssetCopy is the command that copies a build target's assets into its
// output directory (relative to the project), from the Nx assets option: a
// path copied to the same place under the output, or {input, glob, output}.
// ok is false when an entry cannot be copied this way (an ignore list, or a
// path a shell would read).
func AssetCopy(root string, t nxworkspace.Target, outDir string) (line string, ok bool) {
	var assets []json.RawMessage
	if !t.Option("assets", &assets) || len(assets) == 0 {
		return "", false
	}
	var entries []string
	for _, raw := range assets {
		var plain string
		var e struct {
			Input  string   `json:"input"`
			Glob   string   `json:"glob"`
			Output string   `json:"output"`
			Ignore []string `json:"ignore"`
		}
		var input, glob, output string
		whole := 0
		switch {
		case json.Unmarshal(raw, &plain) == nil:
			input, glob, output, whole = ".", relativeTo(root, plain), outDir, 1
		case json.Unmarshal(raw, &e) == nil && e.Input != "" && e.Glob != "" && len(e.Ignore) == 0:
			input, glob, output = relativeTo(root, e.Input), strings.ReplaceAll(e.Glob, "\\", "/"), path.Join(outDir, strings.ReplaceAll(e.Output, "\\", "/"))
		default:
			return "", false
		}
		for _, s := range []string{input, glob, output} {
			if strings.ContainsAny(s, "'\"`$%\\!") {
				return "", false
			}
		}
		entries = append(entries, fmt.Sprintf("['%s','%s','%s',%d]", input, glob, output, whole))
	}
	return `node -e "` + fmt.Sprintf(assetCopyScript, "["+strings.Join(entries, ",")+"]") + `"`, true
}
