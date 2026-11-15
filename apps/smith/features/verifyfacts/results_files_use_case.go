package verifyfacts

import (
	"os"
	"path/filepath"
)

// resultFiles are the test reports smith knows, relative to the project folder,
// in the order it looks. A name with a glob is matched.
var resultFiles = []string{
	"junit.xml", "test-results/junit.xml", "reports/junit.xml", "test-results.xml",
	"TestResults/*.trx", "TestResults/*/*.trx",
	"target/surefire-reports", "build/test-results/test",
}

// FindResults looks for a test report that is already in place in the project
// folder: a JUnit file, a .NET trx, the folder Maven or Gradle write. smith does
// not ask any runner for one.
//
// Parameters: workspaceRoot is the workspace folder; projectRoot is the project
// folder, workspace-relative.
// Returns: the first report, workspace-relative with forward slashes, and true
// when one is in place.
func FindResults(workspaceRoot, projectRoot string) (string, bool) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot))
	for _, pattern := range resultFiles {
		matches, _ := filepath.Glob(filepath.Join(dir, filepath.FromSlash(pattern)))
		for _, m := range matches {
			if _, err := os.Stat(m); err != nil {
				continue
			}
			rel, err := filepath.Rel(workspaceRoot, m)
			if err != nil {
				rel = m
			}
			return filepath.ToSlash(rel), true
		}
	}
	return "", false
}
