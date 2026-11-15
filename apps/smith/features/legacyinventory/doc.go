// Package legacyinventory finds what the old tooling of a Node workspace
// leaves behind once smith is the only tool: the lint, test-runner, build and
// Nx or mnci files, and the devDependencies only they needed. A file or a
// package is reported as removable only when no file that stays imports or
// runs it; when that cannot be shown it is kept, with the reason, so the user
// can decide. It reads a Tree (the files plus the changes a plan has made but
// not written) and changes nothing on disk.
package legacyinventory
