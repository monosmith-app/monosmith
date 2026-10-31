// Package languageserver is `smith lsp`: a Language Server Protocol server on
// stdin and stdout for smith.toml. It reads the schema `smith schema` prints
// and reports a diagnostic for what smith itself would reject, completes keys,
// table names and enumerated values, and shows a key's description on hover.
// It opens no port and runs no task.
package languageserver
