// Package smithfix builds the doctor fixes whose steps are smith commands:
// "smith upgrade --defaults", "smith ci generate". A finding that carries such
// a fix gets a Fix action in every interface (the shell's Doctor tab, the
// editors' Health view), so nobody has to copy a remedy into a terminal.
package smithfix
