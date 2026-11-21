// Package outsidefiles finds the files a project in a sub-folder of a source
// uses outside that sub-folder: a path in a .NET project file that leaves the
// folder (..\LICENSE, a signing key, an imported .props file), a
// Directory.Build.props or .targets above it, and a "files" entry with ../ in
// a package.json. smith import --path takes the sub-folder only, so the build
// of the imported project would not find them. The package reads them before
// the import, and can bring a plain file along into the project folder and
// write its new path in the file that names it.
package outsidefiles
