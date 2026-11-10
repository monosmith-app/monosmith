package releasing

import (
	"fmt"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// notFound reports whether npm said the package or version does not exist
// (E404). An existence probe that finds nothing is an answer, not an error.
func notFound(stderr string) bool {
	return strings.Contains(stderr, "E404") || strings.Contains(stderr, "404 Not Found")
}

// verbose reports whether the person asked to see the tool's own output
// (smith release --verbose sets it for the plugins).
func verbose() bool {
	v := os.Getenv(toolrun.VerboseEnv)
	return v != "" && v != "0"
}

// quietOutput runs npm and returns its stdout. Unlike toolrun.Output it keeps
// stderr back: it is written to the log only when the command fails for a
// reason other than E404, so a probe for a version that is not published yet
// prints nothing.
func quietOutput(dir string, argv []string) (string, error) {
	out, errOut, err := toolrun.Capture(dir, argv)
	if err != nil && !notFound(errOut) {
		_, _ = fmt.Fprint(toolrun.Log, errOut)
	}
	return out, err
}

// quietRun runs an npm command that publishes or tags. On success it prints
// npm's result lines and where it published, and the tarball contents and
// details only with --verbose. On failure it prints everything npm said and
// returns an error that ends with npm's last lines, so the cause is in the
// release's own failure message.
func quietRun(dir string, argv []string) error {
	out, errOut, err := toolrun.Capture(dir, argv)
	if err == nil && !verbose() {
		errOut = keepPublishLines(errOut)
	}
	_, _ = fmt.Fprint(toolrun.Log, out+errOut)
	if err != nil {
		return fmt.Errorf("%w\n%s", err, lastLines(errOut, 8))
	}
	return nil
}

// keepPublishLines drops the "npm notice" lines (the package listing, its
// sizes and digests) and keeps the one that says where it published.
func keepPublishLines(stderr string) string {
	var kept []string
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, "npm notice") && !strings.Contains(line, "Publishing to") {
			continue
		}
		if strings.TrimSpace(line) != "" {
			kept = append(kept, line)
		}
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, "\n") + "\n"
}

func lastLines(text string, n int) string {
	lines := strings.Split(strings.TrimRight(text, "\r\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
