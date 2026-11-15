package convert

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// PruneLockfile drops the packages drop selects from package-lock.json, at
// every depth. npm keeps an optional peer (@mnci/eslint-config's
// @nx/eslint-plugin) for as long as the lockfile lists it, so removing Nx from
// package.json alone leaves Nx installed; without the entries, npm install
// resolves the tree again and drops everything only Nx needed.
func PruneLockfile(raw []byte, drop func(name string) bool) ([]byte, []string, error) {
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("package-lock.json: %w", err)
	}
	packages := doc.Object("packages", false)
	if packages == nil {
		return raw, nil, nil
	}
	var dropped []string
	for _, key := range packages.Keys() {
		i := strings.LastIndex(key, "node_modules/")
		if i < 0 {
			continue
		}
		if drop(key[i+len("node_modules/"):]) {
			packages.Delete(key)
			dropped = append(dropped, key)
		}
	}
	if len(dropped) == 0 {
		return raw, nil, nil
	}
	return doc.Bytes(), dropped, nil
}
