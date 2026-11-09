// Package derivefile applies the three ownership modes of a derived file:
// the whole file with a generated header, a region between smith:begin and
// smith:end markers, and listed keys of a JSON or TOML file. For each mode it
// says what content smith owns (wanted and on disk, so their hashes compare)
// and how to merge the wanted content into the file without touching the rest.
// See ADR 0046.
package derivefile
