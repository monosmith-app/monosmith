package pipeline

import (
	"strings"
)

// AddUse adds a module directory to go.work: inside the first use ( ... )
// block when there is one, else as a use line at the end. Everything else
// stays as it is.
func AddUse(src, root string) string {
	entry := "./" + strings.TrimPrefix(root, "./")
	if loc := useBlock.FindStringSubmatchIndex(src); loc != nil {
		closing := loc[3] // the end of the block's contents, at ")"
		body := src[loc[2]:closing]
		insert := "\t" + entry + "\n"
		if body != "" && !strings.HasSuffix(body, "\n") {
			insert = "\n" + insert
		}
		return src[:closing] + insert + src[closing:]
	}
	if src != "" && !strings.HasSuffix(src, "\n") {
		src += "\n"
	}
	return src + "\nuse " + entry + "\n"
}
