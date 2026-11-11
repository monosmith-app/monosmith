// Package depsrepair is `smith deps repair`: it repairs the installed
// dependencies of a workspace when they are broken. Each language plugin says
// how its tool does each repair kind (prune, clean reinstall, rebuild the
// lockfile, clear the download cache); the person picks one kind; smith shows
// what will be deleted and run, asks, and then deletes the paths, runs the
// commands and installs again.
//
// The plugin only describes (dependencies/repair). This package does the
// deleting, and it deletes a link as a link: it never follows a symlink or a
// Windows junction into the folder it points to. It never commits.
package depsrepair
