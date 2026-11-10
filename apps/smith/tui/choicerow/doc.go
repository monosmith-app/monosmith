// Package choicerow draws the rows of a choice the same way on every screen:
// the tick box "[ ]", "[x]" or "[-]", the radio mark "( )" and a row with its
// cursor mark. A screen calls Box, Radio, Cursor or Render, and never spells a
// box or a cursor mark by hand. Guard tests fail on a new hand-drawn box or
// cursor mark.
package choicerow
