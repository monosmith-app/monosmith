// Package tsconfigchain owns the TypeScript configuration chain of a workspace:
// every tsconfig.json smith writes reaches the root tsconfig.base.json, directly
// or through the root tsconfig.browser.json, and the base extends the
// @monosmith/tsconfig preset. One edit of the base then reaches every project.
//
// It writes the root files, says how far a project is from the root, and
// follows the extends chain of a project to see whether it reaches the base.
package tsconfigchain
