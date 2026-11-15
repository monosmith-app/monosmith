package doctorfix

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// StepCommand is a fix step as a person reads it: the environment the step adds
// ("GOBIN=/work/.smith/bin"), then the command.
func StepCommand(s protocol.FixStep) string {
	return strings.TrimSpace(strings.Join(s.Env, " ") + " " + displayCommand(s.Argv))
}

// stepEnv turns a step's KEY=VALUE entries into the map the process runner takes.
func stepEnv(s protocol.FixStep) map[string]string {
	if len(s.Env) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range s.Env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			out[k] = v
		}
	}
	return out
}

// displayCommand is a step's command line as a person reads it: the program
// smith runs is written "smith", not the path of the binary that happens to
// run on this machine. Other programs keep their names.
func displayCommand(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	shown := append([]string(nil), argv...)
	// A Windows path keeps its backslashes on other systems, so split on both.
	base := strings.ToLower(shown[0])
	base = strings.TrimSuffix(base[strings.LastIndexAny(base, `/\`)+1:], ".exe")
	if base == "smith" || strings.HasPrefix(base, "smith-") {
		shown[0] = "smith"
	}
	return strings.Join(shown, " ")
}
