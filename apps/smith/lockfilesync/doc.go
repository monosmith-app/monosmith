// Package lockfilesync leaves a Node workspace's package-lock.json in step
// with its package.json after smith has edited the manifests: one npm
// install, a check with npm ci --dry-run, a second install when the check
// fails, and, as a last resort, a lockfile rebuilt from scratch when git can
// take that back. Adoption and conversion both go through Sync; the hidden
// `smith npm-lockfile` runs it as the closing step of a full conversion.
package lockfilesync
