package doctorfix

import (
	"path/filepath"
	"strings"
)

// displayCommand is a step's command line as a person reads it: the program
// smith runs is written "smith", not the path of the binary that happens to
// run on this machine. Other programs keep their names.
func displayCommand(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	shown := append([]string(nil), argv...)
	base := strings.TrimSuffix(strings.ToLower(filepath.Base(shown[0])), ".exe")
	if base == "smith" || strings.HasPrefix(base, "smith-") {
		shown[0] = "smith"
	}
	return strings.Join(shown, " ")
}
