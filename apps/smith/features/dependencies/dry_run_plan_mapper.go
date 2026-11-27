package dependencies

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// PrintDryRun writes what smith up would change for the updates chosen, and
// changes nothing: each declaration it would rewrite (the manifest, the
// package and the spec before and after), then whether it would install, and
// that nothing was written.
//
// Parameters: w receives the text; chosen are the updates up would apply;
// install says whether up would install after it rewrote the manifests.
func PrintDryRun(w io.Writer, chosen []Update, install bool) {
	plan := rewrites(chosen)
	fmt.Fprintln(w, words.Count("dependencies.up.dry_run.head", len(plan)))
	for _, r := range plan {
		fmt.Fprintln(w, words.Say("dependencies.up.dry_run.site", "file", r.Declaration.File, "name", r.Declaration.Name, "from", r.Declaration.Spec, "to", r.Spec))
	}
	if install {
		fmt.Fprintln(w, words.Say("dependencies.up.dry_run.install"))
	}
	fmt.Fprintln(w, words.Say("dependencies.up.dry_run.nothing"))
}
