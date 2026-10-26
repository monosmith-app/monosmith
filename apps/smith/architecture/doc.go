// Package architecture is the architecture a workspace follows: how its code
// is laid out. The known architectures are listed in one place (the catalog),
// the one a workspace uses is read from its smith.toml, and choosing another
// edits that file, the same way `smith config set` does: only the table that
// turns an architecture on is added or removed. The Settings tab of the
// terminal interface and `smith architecture` both go through here.
package architecture
