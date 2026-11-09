package siblingfiles

import "strings"

// MinSiblings is the least number of projects that must hold a file in the same
// shape before the shape counts as the workspace's convention.
const MinSiblings = 3

// isShared reports whether n files in the same shape, among total files at the
// same path, make the shape a convention: at least MinSiblings, and at least two
// thirds of the files.
func isShared(n, total int) bool {
	return n >= MinSiblings && n*3 >= total*2
}

// IsConfigFile reports whether a file name in a project root is a config file
// that can be shared: a tsconfig, an swc config, or the config of Jest, Vitest,
// ESLint, Prettier or Babel. A manifest (package.json), a test setup file and a
// source file are never config files, so they are never derived.
func IsConfigFile(name string) bool {
	switch {
	case strings.HasPrefix(name, "tsconfig") && strings.HasSuffix(name, ".json"):
		return true
	case strings.HasSuffix(name, ".swcrc"), name == ".babelrc":
		return true
	}
	for _, prefix := range []string{"jest.config.", "vitest.config.", "eslint.config.", "babel.config.", ".eslintrc", ".prettierrc"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
