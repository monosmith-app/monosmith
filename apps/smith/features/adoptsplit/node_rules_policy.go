package adoptsplit

// nodeRules: formatter, linter and commit-tool configuration stays at the
// root, where the tools find it for every project. The ESLint files stay too:
// Split renames a legacy .eslintrc to .pre-smith, since the workspace uses
// the shared flat config. package.json and tsconfig.json move with the
// project and Split cuts them. jest.config.* moves whole: its paths are
// relative to the project (rootDir). The lockfiles stay: with npm
// workspaces the one lockfile lives at the root and npm install rewrites it.
var nodeRules = Rules{
	Ecosystem: Node,
	Workspace: []string{
		".prettierrc*", "prettier.config.*", ".prettierignore",
		".eslintrc*", "eslint.config.*", ".eslintignore",
		".stylelintrc*", "stylelint.config.*", ".stylelintignore",
		".lintstagedrc*", "lint-staged.config.*",
	},
	Lockfiles: []string{"yarn.lock", "pnpm-lock.yaml", "bun.lockb", "bun.lock", "package-lock.json", "npm-shrinkwrap.json"},
	Sort:      SortNode,
	Split:     SplitNode,
}
