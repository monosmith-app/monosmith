// Package adoptmove moves the one project of a plain repository (a package.json,
// go.mod, pyproject.toml or Cargo.toml at the root) into a folder of its own,
// because smith needs one folder per project. It finds the project, derives its
// name, offers the folders it may go in, works out which files belong to the
// project and which to the repository, and moves them with git mv. smith never
// chooses the folder: the caller asks the person (ADR 0070).
package adoptmove
