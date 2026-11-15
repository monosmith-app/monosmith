// Package cachesetup is the guided setup of the remote cache (`smith cache
// setup`): which storage, where in it, the key, the backend's credential and
// the steps of each CI provider. Every decision is a choice with its options
// explained. smith runs a command, writes a file or stores a secret only after
// the person confirms, and shows the exact command first. The cloud
// command-line programs (az, aws) are reached through a port, so tests use a
// fake and no test calls a cloud.
package cachesetup
