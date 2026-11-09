// Package words gives an interface the text for a key: Say fills {name}
// placeholders by name and Count picks the plural form. The texts are the catalogue
// in libs/words/catalog, embedded in the binary; English is the reference and a
// locale that lacks a key shows the English text, key by key. The module's
// documentation (libs/words/doc.go) tells how the surfaces and the generator fit.
package words
