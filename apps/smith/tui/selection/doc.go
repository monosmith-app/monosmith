// Package selection is how text is selected with the mouse while mouse
// reporting is on. A screen lists the Regions a drag may use (Divide cuts the
// panels apart along their dividers). State then does the rest: it starts at a
// press, follows the drag inside the region it began in, and on the release
// keeps the highlight. Text gives back the plain text of the selected cells and
// Paint draws the highlight over the screen rows. A screen never selects text
// by hand.
package selection
