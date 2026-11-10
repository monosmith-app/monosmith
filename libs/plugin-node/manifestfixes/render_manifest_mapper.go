package manifestfixes

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
)

// render writes a package.json the way it was read: with the line endings it
// had, so the edit shows in a diff as the lines it changed.
func render(doc *jsondoc.Document, original []byte) string {
	return lineendings.Match(string(original), string(doc.Bytes()))
}
