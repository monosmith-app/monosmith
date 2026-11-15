package releasepreview

import (
	"fmt"
	"io"
	"os"
)

// Post puts the preview on the pull request this CI run builds. Outside a
// pull request it says so and does nothing; a failure is reported, never
// fatal, because a missing comment must not fail the pull request.
func Post(markdown string, out io.Writer) {
	pr, ok, why := FromEnvironment(os.Getenv)
	if !ok {
		fmt.Fprintf(out, "Release preview not posted: %s.\n", why)
		return
	}
	done, err := Upsert(pr, Marker+"\n"+markdown)
	if err != nil {
		fmt.Fprintf(out, "warning: could not post the release preview on pull request %d: %v\n", pr.Number, err)
		return
	}
	fmt.Fprintf(out, "Release preview %s on pull request %d.\n", done, pr.Number)
}
