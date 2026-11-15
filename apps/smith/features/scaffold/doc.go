// Package scaffold creates workspaces (smith new) and projects (smith add). A
// project is one of three kinds — an app (deployed, never published), a lib
// (internal, never published) or a package (published) — in any language a
// plugin provides; the plugin writes the files, smith picks the directory and
// regenerates the pipelines.
package scaffold
