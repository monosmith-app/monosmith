package dialect

import (
	"fmt"
	"io"
	"sort"
)

// Exemplar makes the dialect of one chosen project (smith add --like): its
// values for the keys a plugin of the new project's stack follows are sent to
// the plugins, and the report says what was taken from it.
//
// Parameters: name is the chosen project; shape is the shape it reports; kind
// words the report; follows are the keys the stack applies; out gets the report.
// Returns: the dialect to follow. It prints nothing when shape is empty.
func Exemplar(name string, shape map[string]string, kind string, follows map[string]bool, out io.Writer) Dialect {
	d := Dialect{kind: kind, expected: map[string]expectation{}}
	keys := make([]string, 0, len(shape))
	for k := range shape {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		fmt.Fprintf(out, "Following %s (--like):\n", name)
	}
	for _, k := range keys {
		d.expected[k] = expectation{shape[k], 1, 1}
		if follows[k] {
			if d.Shape == nil {
				d.Shape = map[string]string{}
			}
			d.Shape[k] = shape[k]
			fmt.Fprintf(out, "  %s = %s\n", k, shape[k])
		} else {
			fmt.Fprintf(out, "  %s = %s (no plugin of the stack follows this key)\n", k, shape[k])
		}
	}
	return d
}
