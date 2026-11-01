package dialect

import "fmt"

// Summary lists, one line each, the values the siblings agree on, for a plan
// shown before anything is written.
//
// Parameters: siblings are the shapes of the projects of the same kind and
// language; kind words the lines.
// Returns: the lines; empty when the siblings have no habit.
func Summary(siblings []map[string]string, kind string) []string {
	var out []string
	for _, f := range CountFacts(siblings) {
		if top, ok := f.Dominant(); ok {
			out = append(out, fmt.Sprintf("%s = %s (%d of %d %ss)", f.Key, top.Value, top.N, f.Total, kind))
		}
	}
	return out
}
