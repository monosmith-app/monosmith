package manifestfixes

import (
	"bytes"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// render writes a package.json the way it was read: with the line endings it
// had, so the edit shows in a diff as the lines it changed.
func render(doc *jsondoc.Document, original []byte) string {
	out := doc.Bytes()
	if bytes.Contains(original, []byte("\r\n")) {
		out = bytes.ReplaceAll(out, []byte("\n"), []byte("\r\n"))
	}
	return string(out)
}
