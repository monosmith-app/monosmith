package protocol

import (
	"fmt"
	"regexp"
	"slices"
)

// choiceName is what an axis or a choice may be called: lowercase letters,
// digits and dashes, so it is a valid bare key in smith.toml and a word in a
// command line.
var choiceName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ValidateChoice checks that a choice is well formed: an axis, a name and a
// language in the plain form, only the scaffold and toolchain roles, and no
// choice that requires or conflicts with itself.
//
// It returns an error that names the choice and the field.
func ValidateChoice(c Choice) error {
	for field, value := range map[string]string{"axis": c.Axis, "name": c.Name, "language": c.Language} {
		if !choiceName.MatchString(value) {
			return fmt.Errorf("choice %q: %s %q must be lowercase letters, digits and dashes", c.Name, field, value)
		}
	}
	for _, role := range c.Roles {
		if role != RoleScaffold && role != RoleToolchain {
			return fmt.Errorf("choice %s: role %q must be scaffold or toolchain", c.Name, role)
		}
	}
	for _, list := range []struct {
		field string
		names []string
	}{{"requires", c.Requires}, {"conflicts", c.Conflicts}} {
		if slices.Contains(list.names, c.Name) {
			return fmt.Errorf("choice %s lists itself in %s", c.Name, list.field)
		}
		for _, n := range list.names {
			if !choiceName.MatchString(n) {
				return fmt.Errorf("choice %s: %s %q must be the name of a choice", c.Name, list.field, n)
			}
		}
	}
	for _, axis := range c.Needs {
		if !choiceName.MatchString(axis) {
			return fmt.Errorf("choice %s: needs %q must be the name of an axis", c.Name, axis)
		}
	}
	return nil
}
