package lintfix

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
)

// ESLintNotInstalled is the cause reported for a project when the workspace
// has no ESLint.
const ESLintNotInstalled = "ESLint is not installed in the workspace"

// ESLintRemedy says what to do about an ESLint that could not run in a
// project. No ESLint installed: install. One older than smith's config needs:
// raise it. Else: read the whole error from the plain lint.
//
// Parameters: project is the project's name, installed the ESLint version in
// the workspace ("" when none).
func ESLintRemedy(project, installed string) string {
	switch {
	case installed == "":
		return "run npm install, then " + LintCommand(project)
	case !eslintpeer.Meets(installed):
		return fmt.Sprintf("smith's config needs eslint %s and the workspace has %s: set \"eslint\": %q in package.json, then run npm install", eslintpeer.Range, installed, eslintpeer.Range)
	}
	return "run " + LintCommand(project) + " to see the whole error"
}
