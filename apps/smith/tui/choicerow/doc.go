// Package choicerow draws the rows of a choice the same way on every screen:
// the tick box "[ ]", "[x]" or "[-]", the radio mark "( )" and a row with its
// cursor mark. A screen calls Box, Radio or Render, and never spells a box by
// hand. The button package's guard test fails on a new hand-drawn box.
package choicerow
