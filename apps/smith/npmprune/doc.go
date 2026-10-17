// Package npmprune is `smith npm-prune`, the node plugin's prune task: it
// writes a Node app's deploy output (its production package.json and
// lockfile, and the workspace packages it uses in workspace_modules/) into
// the app's build directory. Hidden: the task calls it through {smith},
// people run smith run prune.
package npmprune
