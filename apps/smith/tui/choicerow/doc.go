// Package choicerow is the one pick-list and tick-box row of the TUI. Every
// list of choices, of answers, of options and of tick boxes draws its rows with
// Lines or Line (a Row: cursor mark, lead, label, hint), so they share the
// glyphs ("[ ]", "[x]", "[-]", "( )", the cursor mark), the selected look, the
// row hover look of the theme and the muted look of a row that cannot be
// chosen. A lone tick box outside a list is Toggle. At maps a click or a hover
// to a row that may wrap over several lines. A screen never spells a box or a
// cursor mark by hand: guard tests fail on a new one.
package choicerow
