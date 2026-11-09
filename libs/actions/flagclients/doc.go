// Package flagclients turns the flag vocabulary (package flagvocabulary) into the
// files the clients read: the constants of the VS Code extension, the constants of
// the JetBrains plugin and the reference table in the docs. The functions are
// pure; command flaggen writes the files and a test fails when one is stale.
package flagclients
