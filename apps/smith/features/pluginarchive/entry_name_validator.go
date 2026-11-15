package pluginarchive

import (
	"fmt"
	"strings"
)

// ValidateName accepts a relative, slash-separated path that stays inside the
// directory it is extracted into on every operating system: no absolute
// path, drive letter, backslash, "." or ".." element, empty element, control
// character, or Windows device name.
func ValidateName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("an entry has no name")
	case strings.ContainsAny(name, "\\:"):
		return fmt.Errorf("unsafe entry name %q: backslash or colon", name)
	case strings.HasPrefix(name, "/"):
		return fmt.Errorf("unsafe entry name %q: absolute path", name)
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("unsafe entry name %q: control character", name)
		}
	}
	for _, part := range strings.Split(name, "/") {
		switch {
		case part == "" || part == "." || part == "..":
			return fmt.Errorf("unsafe entry name %q: path traversal or empty element", name)
		case strings.HasSuffix(part, ".") || strings.HasSuffix(part, " "):
			return fmt.Errorf("unsafe entry name %q: element ends with a dot or space", name)
		case isWindowsDevice(part):
			return fmt.Errorf("unsafe entry name %q: reserved device name", name)
		}
	}
	return nil
}

func isWindowsDevice(part string) bool {
	base := strings.ToUpper(part)
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '0' && base[3] <= '9' {
		return true
	}
	return false
}
