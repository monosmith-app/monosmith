package declaredimports

import "fmt"

// UndeclaredImport is one package a module's file imports from a module its
// go.mod does not require.
type UndeclaredImport struct {
	// File is the importing file, relative to the module.
	File string
	// Package is the imported package's path.
	Package string
}

func (u UndeclaredImport) String() string {
	return fmt.Sprintf("%s imports %s, which go.mod does not require", u.File, u.Package)
}
