// Package legacyjesthelpers finds the Jest helper files that older versions of
// smith convert wrote at the workspace root (jest.resolver.cjs,
// jest.asset-transform.cjs, and a jest.preset.js with the preset values
// written out) and plans the cleanup: the same behaviour now comes from
// @monosmith/jest-preset, so the files go, the configs that named them name the
// package, and the package becomes a root devDependency. smith doctor offers
// the plan as a fix.
package legacyjesthelpers
