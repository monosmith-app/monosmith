// Package statuspaint colours the glyph of the status lines a command prints for
// a person (check, cross, warning, skipped, start), the way the log palette of the
// cli-output-design skill says: green for passed, red for failed, yellow for a
// warning, dim for what did not run. Only the glyph is painted, the words stay as
// they are, and the decision is the one every table of smith follows
// (termcolour.Wanted): a terminal, or a CI that renders colour, and never
// NO_COLOR or TERM=dumb. A host that reads the lines gets them plain.
package statuspaint
