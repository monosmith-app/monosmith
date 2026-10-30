package declaredimports

import "fmt"

// UndeclaredImport is one module a project's file imports without declaring
// the distribution that provides it where it must be.
type UndeclaredImport struct {
	// File is the importing file, relative to the project.
	File string
	// Module is the imported top-level module.
	Module string
	// DevOnly is true when a development dependency provides the module, which
	// installing the package does not install.
	DevOnly bool
}

func (u UndeclaredImport) String() string {
	if u.DevOnly {
		return fmt.Sprintf("%s imports %s, declared only in development dependencies", u.File, u.Module)
	}
	return fmt.Sprintf("%s imports %s, which pyproject.toml does not declare", u.File, u.Module)
}
