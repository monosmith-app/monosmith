// Package words is where every sentence, label and hint the interface shows
// lives. A screen asks for its text by key (words.Say("doctor.headline.passed"))
// and never writes the text itself, so one place holds all the wording: to
// reword the interface or translate it, change the catalog, not the code.
//
// The catalog is one JSON file per screen under catalog/<locale>/, each a flat
// map from key to text. A key is the screen, then what the text is
// (doctor.headline.failed). Text takes {name} placeholders, filled by name so a
// translation can put them in any order, and a count's plural forms are two
// keys, key.one and key.other.
//
// English is the reference: a locale that lacks a key shows the English text.
// The guard test in this package fails a key that no code uses, a key that the
// code asks for and the catalog lacks, a placeholder the call does not fill,
// and raw wording in a package that has moved to the catalog.
package words
