package flagclients

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

// Outputs are the generated files, relative to the repository root.
var Outputs = []string{
	"apps/vscode/src/smith-flags/smith-flags.config.ts",
	"apps/jetbrains/src/main/kotlin/dev/monosmith/jetbrains/smithcommand/SmithFlagsConfig.kt",
	"docs/reference/flag-vocabulary.md",
}

// Render returns the text of each generated file, in the order of Outputs.
func Render() []string {
	concepts := flagvocabulary.All()
	return []string{TypeScript(concepts), Kotlin(concepts), Markdown(concepts, flagvocabulary.Exceptions())}
}

// Sync compares the generated files under root with what the vocabulary says and
// returns the paths that differ or are missing. Unless check is true it writes
// them. Line endings do not count: a checkout may have CRLF.
//
// Parameters: root is the repository root; check says to write nothing.
// Errors: a file that cannot be read or written.
func Sync(root string, check bool) (stale []string, err error) {
	for i, text := range Render() {
		path := filepath.Join(root, filepath.FromSlash(Outputs[i]))
		current, readErr := os.ReadFile(path)
		if readErr == nil && bytes.Equal(bytes.ReplaceAll(current, []byte("\r\n"), []byte("\n")), []byte(text)) {
			continue
		}
		stale = append(stale, Outputs[i])
		if check {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return stale, err
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			return stale, err
		}
	}
	return stale, nil
}
