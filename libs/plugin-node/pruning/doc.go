// Package pruning makes a Node app's deploy output: its build directory
// with a package.json and lockfile of its production tree only, and the
// workspace packages it uses copied into workspace_modules/, so the
// directory installs on its own (npm ci --omit=dev), in a container image
// for example. It does what Nx's prune, prune-lockfile and
// copy-workspace-modules did.
package pruning
