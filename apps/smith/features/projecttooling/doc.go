// Package projecttooling is the per-project step "keep its own build and
// lint tools, or use smith's" (owner decision D3 of the any-repository
// review, issue #553) for a Node (JavaScript or TypeScript) project.
//
// The step has three calls, each usable alone:
//
//   - Assess reads one project and says which of its tools are legacy (an
//     ESLint 8 .eslintrc, create-react-app, a tsc build into lib/, a rollup or
//     webpack build), and so which answer smith recommends.
//   - Decide settles the answer of each project: a flag (--tooling) answers;
//     in a terminal smith asks and recommends; with --yes the project keeps its
//     own tools and a line says how to choose smith's. It never picks smith's
//     unasked.
//   - Apply takes the answer for one project. With "smith" it first puts
//     smith's tools in place (the root eslint.config.mjs with
//     @monosmith/eslint-config, a lint script, smith's build), then retires the
//     old tools through the full conversion's removal plan (packages
//     legacyinventory, removalplan and fullconversion), and only those whose
//     replacement is in place. What smith cannot replace is a named item of the
//     result's Left list.
//
// Apply takes a small input (the workspace root, the project folder and the
// answer), so `smith adopt` and `smith import` call the same step. Python and
// .NET projects are out of scope: Assess reports them as not Node.
package projecttooling
