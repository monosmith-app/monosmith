// Package turboplan decides how a Turborepo workspace becomes a smith
// workspace: its packages become projects, turbo.json's tasks become task
// settings in smith.toml, and pnpm or yarn become the npm workspaces that
// smith install uses. It reports everything it cannot map and changes
// nothing on disk.
package turboplan
