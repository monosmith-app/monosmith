// Package outputignore makes sure git ignores what a language's tools write
// into a project folder (bin/ and obj/ of .NET, node_modules/ of Node), so a
// commit of a project that came into the workspace never takes its build
// output. The language plugins name the patterns (InitializeResult.Ignores);
// this package finds the ones git does not ignore yet in one folder and adds
// them to the workspace .gitignore, anchored to that folder.
package outputignore
