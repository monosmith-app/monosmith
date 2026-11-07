// Package marks holds the status symbols smith prints in the CLI, the CI log
// and the interactive shell: passed, failed, warning, skipped, cancelled and cached.
//
// The symbols were chosen because they draw in one cell in every monospace
// font and have no emoji form, so no terminal replaces them with a wide colour
// picture that overlaps the next character. The marks package depends on
// nothing, so any slice may import it.
package marks
