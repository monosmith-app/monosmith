package cipipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"

// preserveFile is Preserve for the pipeline file at rel (slash-separated, from
// root) whose current text is old: it keeps what Preserve keeps, and gives the
// result the line ending of the file that is there, else the one the
// repository uses. A pipeline in a CRLF checkout is therefore not rewritten
// with LF and does not show as out of date.
func preserveFile(root, rel, old, rendered string) (content string, lost []string) {
	content, lost = Preserve(lineendings.Convert(old, false), rendered)
	return lineendings.Resolve(root, rel, content), lost
}
