// Package pluginadmit lets an external plugin start only once it is trusted:
// it checks each plugin the workspace needs, asks the person in a terminal
// when there is one, and otherwise refuses with the command that fixes it.
// It also describes what each external plugin may see and whether it is
// trusted, for `smith plugin list` and `smith doctor`.
package pluginadmit
