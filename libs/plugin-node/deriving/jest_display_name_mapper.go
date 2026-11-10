package deriving

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// JestDisplayName is the id of the renderer of a Jest config's displayName.
const JestDisplayName = "node.jest.display-name"

var jestConfigs = []string{"jest.config.ts", "jest.config.js", "jest.config.cjs", "jest.config.mjs"}

// jestDisplayName derives the displayName line of a project's Jest config: the
// package name. The renderer owns the lines between `// smith:begin
// node.jest.display-name` and `// smith:end node.jest.display-name`, so the
// config says where the line goes; a config without the markers gets nothing.
func jestDisplayName(workspaceRoot string, p protocol.Project) (protocol.DerivedFile, bool) {
	for _, name := range jestConfigs {
		rel := path.Join(p.Root, name)
		content, ok := readFile(workspaceRoot, rel)
		if !ok {
			continue
		}
		indent, found := markerIndent(string(content))
		if !found {
			return protocol.DerivedFile{}, false
		}
		display := p.Name
		if len(p.Exports) > 0 {
			display = p.Exports[0]
		}
		quoted := "'" + strings.ReplaceAll(strings.ReplaceAll(display, `\`, `\\`), "'", `\'`) + "'"
		return protocol.DerivedFile{
			Project: p.Name, Path: rel, Mode: protocol.DeriveRegion, Comment: "//",
			Renderer: JestDisplayName, Version: 1, Content: indent + "displayName: " + quoted + ",\n",
		}, true
	}
	return protocol.DerivedFile{}, false
}

// markerIndent is the indentation of the begin marker line, which the derived
// lines share.
func markerIndent(content string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		if i := strings.Index(line, "smith:begin "+JestDisplayName); i >= 0 {
			return line[:len(line)-len(strings.TrimLeft(line, " \t"))], true
		}
	}
	return "", false
}
