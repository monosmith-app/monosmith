// Package toolcatalog is the data of the programs the language plugins need
// (git, Python, the .NET SDK, Docker): what each is for, how to detect it, and
// how each package manager gets it. The Go plugins publish entries with For in
// InitializeResult.Tools; a plugin in another language writes the same fields
// by hand. Every entry carries the page it was checked against and the date.
package toolcatalog
