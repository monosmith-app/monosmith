// Package convertchoices is the set of decisions behind `smith convert --level
// full`: what lints, what runs the tests, what builds, and what runs CI once
// smith is the only tool. It derives each default from what the workspace
// already uses, checks the values the flags give, asks the ones left open as
// numbered questions in a terminal, and spells a set of choices back as the
// flags that make it.
package convertchoices
