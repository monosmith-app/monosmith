// Package flagvocabulary is the one declaration of the flags that several smith
// commands share: for each concept (the project to act on, the language, the
// output path, "do not ask") its canonical name, shorthand, synonyms, value type,
// meaning and the commands that take it. It is DATA (flag_vocabulary.json), read
// by the command line to register the flags, by the generator that writes the
// constants of the VS Code and JetBrains clients and the reference table, and by
// the test that walks the command tree. No client keeps a flag name of its own.
// The decision is ADR 0084: one concept, one flag; add the concept here first.
package flagvocabulary
