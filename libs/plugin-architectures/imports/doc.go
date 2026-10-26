// Package imports reads which of a project's own folders each source file
// imports, in every language the architecture check supports. Imports of
// libraries and of the language's own packages are left out: only code of
// the project itself can break its layout.
package imports
