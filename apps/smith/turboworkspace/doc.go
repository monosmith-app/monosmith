// Package turboworkspace reads a Turborepo workspace from disk: turbo.json
// (and the turbo.json some packages keep beside their package.json), the root
// package.json, pnpm-workspace.yaml and every package the workspace lists. It
// runs nothing, so a conversion needs no installed dependencies.
package turboworkspace
