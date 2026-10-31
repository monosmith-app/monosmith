// Package dialect lets a new project follow what its sibling projects do.
// The language plugin describes each project as a shape (build tool, test
// runner, where the entry points lead, tsconfig files); the package counts
// the shapes of the projects of one kind and language, reports what they
// agree on with the reason, asks when they split and warns when the new
// project does not match. It knows no key and no language: a shape is a map
// of strings (ADR 0040).
package dialect
