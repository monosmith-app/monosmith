package toolchainpin

import (
	"regexp"
	"strings"
)

var versionNumber = regexp.MustCompile(`\d+(?:\.\d+)*`)

// Installed returns the label and version number of a probed toolchain
// ("node 20.11.0" gives "node" and "20.11.0"); both are empty when it names
// none.
func Installed(toolchain string) (label, version string) {
	label, rest, _ := strings.Cut(strings.TrimSpace(toolchain), " ")
	if v := versionNumber.FindString(rest); v != "" {
		return label, v
	}
	return "", ""
}

// Satisfies reports whether an installed version meets a pin. A pin is a
// version prefix on dot boundaries: "22" accepts 22.3.0 and not 220.0.0,
// "1.23" accepts 1.23.4 and not 1.230.0, "1.23.4" accepts only itself. A
// leading "v" is ignored.
func Satisfies(pin, installed string) bool {
	pin = strings.TrimPrefix(strings.TrimSpace(pin), "v")
	if pin == "" {
		return true
	}
	return installed == pin || strings.HasPrefix(installed, pin+".")
}
