// Package adoptdecision settles the choices smith adopt --init used to make
// in silence: the line-ending file, the agent files, the CI provider and what
// happens to a pipeline the repository already has. Each is a question in a
// terminal, a flag everywhere else, and a printed line that says what was
// chosen and how to change it. Without a terminal and without --yes the
// package refuses and names the flags.
package adoptdecision
