package tooldefaults

import "strings"

// WithDefaultInputs returns inputs plus every smith default named in args (the
// placeholder form "{workspaceRoot}/.smith/defaults/<file>" that WriteDefault's
// callers pass to tools). The files are git-ignored, so the task cache reads
// them from disk by name: editing a default, or a release that changes one,
// then changes the key of the tasks that use it.
func WithDefaultInputs(inputs, args []string) []string {
	out := append([]string(nil), inputs...)
	prefix := "{workspaceRoot}/" + Dir + "/"
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			out = append(out, a)
		}
	}
	return out
}
