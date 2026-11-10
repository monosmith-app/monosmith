package adoptsplit

import "strings"

// nodeToolNames are the devDependencies that are tools, by exact name.
var nodeToolNames = map[string]bool{
	"typescript": true, "ts-node": true, "tsx": true, "nodemon": true, "rimraf": true,
	"prettier": true, "husky": true, "lint-staged": true,
	"jest": true, "ts-jest": true, "babel-jest": true, "vitest": true, "mocha": true, "nyc": true, "c8": true,
	"npm-run-all": true, "npm-run-all2": true, "cross-env": true, "concurrently": true,
}

// nodeToolPrefixes are the families of tools: formatter plugins, test-runner
// packages, commit tools, and the type packages the compiler reads.
var nodeToolPrefixes = []string{"@types/", "@commitlint/", "prettier-", "@jest/", "jest-", "@vitest/"}

// IsNodeTool says whether a devDependency is a tool of the workspace: a
// compiler, formatter, test runner, or something that belongs to one. A name
// not listed is not a tool, and the person is asked about it. The ESLint
// stack is not listed here: see IsNodeLint.
//
// Parameters: name is the package name.
func IsNodeTool(name string) bool {
	if nodeToolNames[name] {
		return true
	}
	for _, prefix := range nodeToolPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// IsNodeLint says whether a devDependency is part of the ESLint stack that
// the workspace's shared lint config (@monosmith/eslint-config) replaces: it
// brings its own ESLint, parser and plugins, and a second set of them would
// not install beside it (the shared config needs the current ESLint).
//
// Parameters: name is the package name.
func IsNodeLint(name string) bool {
	if name == "eslint" || name == "typescript-eslint" {
		return true
	}
	for _, prefix := range []string{"@typescript-eslint/", "@eslint/", "eslint-"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
