package languages

import (
	"regexp"
	"strings"
)

var (
	pyRoleFile = regexp.MustCompile(`^` + snakeName + `_(?:` + roleAlternation + `)\.py$`)
	pyTestFile = regexp.MustCompile(`^test_[a-z0-9_]+\.py$`)
)

// pythonRules: a package's top folder holds only __init__.py and
// __main__.py; a slice module is <snake>_<role>.py; a test is
// test_<name>.py; packages are snake_case. Modules outside a package
// (setup.py, scripts) are not checked.
var pythonRules = Rules{
	Name:   "python",
	Roots:  []string{"src", ""},
	Inputs: []string{"**/*.py"},
	Skip:   func(name string, _ Dir) bool { return strings.HasSuffix(name, ".egg-info") },
	Folder: func(name string, in Dir) string {
		if pythonPackage(in) && !snakeCase.MatchString(name) {
			return "a Python package is snake_case: its name is an import path"
		}
		return ""
	},
	Source: func(name string, in Dir) bool {
		return strings.HasSuffix(name, ".py") && (pythonPackage(in) || in.Rel != "")
	},
	File: func(name string, in Dir) string {
		return PythonFile(name, pythonPackage(in) && !insidePackage(in))
	},
	Production: func(name string, _ Dir) bool {
		return name != "__init__.py" && name != "conftest.py" && !pyTestFile.MatchString(name)
	},
}

func pythonPackage(d Dir) bool { return d.Files["__init__.py"] }

// insidePackage reports whether a folder is within a package (so it is not
// the package's top folder).
func insidePackage(d Dir) bool {
	for p := d.Parent; p != nil; p = p.Parent {
		if pythonPackage(*p) {
			return true
		}
	}
	return false
}

// PythonFile says why a Python module breaks the rules, or "" when it does
// not.
func PythonFile(name string, atPackageRoot bool) string {
	switch {
	case name == "__init__.py", name == "conftest.py", pyTestFile.MatchString(name):
		return ""
	case name == "__main__.py" && atPackageRoot:
		return ""
	case atPackageRoot:
		return "only __init__.py and __main__.py sit at the top of a package; move this into a slice package"
	case pyRoleFile.MatchString(name):
		return ""
	}
	return "a Python module is <snake>_<role>.py (roles: " + strings.Join(Roles, ", ") + ")"
}
