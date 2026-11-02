package linking

import (
	"fmt"
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// dependencySections are the places a package.json can name a dependency.
var dependencySections = []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies"}

// AddDependency writes name with spec into the dependencies of a package.json
// text, keeping the other keys and their order, and the keys of dependencies in
// alphabetical order as npm writes them. It returns the new text and false when
// the file already lists the name in dependencies with that spec.
func AddDependency(manifest []byte, name, spec string) ([]byte, bool, error) {
	doc, err := jsondoc.ParseDocument(manifest)
	if err != nil {
		return nil, false, err
	}
	deps := doc.Object("dependencies", true)
	if current, ok := deps.String(name); ok && current == spec {
		return manifest, false, nil
	}
	deps.SetString(name, spec)
	keys := deps.Keys()
	sort.Strings(keys)
	values := map[string]string{}
	for _, k := range keys {
		v, ok := deps.String(k)
		if !ok {
			return nil, false, fmt.Errorf("dependencies.%s is not a string", k)
		}
		values[k] = v
	}
	for _, k := range keys {
		deps.Delete(k)
	}
	for _, k := range keys {
		deps.SetString(k, values[k])
	}
	return doc.Bytes(), true, nil
}

// RemoveDependency takes name out of every dependency section of a
// package.json text, and drops a section it empties. It returns the new text
// and the sections the name was in, none when the file did not list it.
func RemoveDependency(manifest []byte, name string) ([]byte, []string, error) {
	doc, err := jsondoc.ParseDocument(manifest)
	if err != nil {
		return nil, nil, err
	}
	var from []string
	for _, section := range dependencySections {
		if !doc.Has(section) {
			continue
		}
		deps := doc.Object(section, false)
		if deps == nil || !slices.Contains(deps.Keys(), name) {
			continue
		}
		deps.Delete(name)
		from = append(from, section)
		if deps.Len() == 0 {
			doc.Delete(section)
		}
	}
	if len(from) == 0 {
		return manifest, nil, nil
	}
	return doc.Bytes(), from, nil
}

// DeclaredSpec returns the spec a package.json text gives name in any
// dependency section, and false when it does not list it.
func DeclaredSpec(manifest []byte, name string) (string, bool) {
	doc, err := jsondoc.ParseDocument(manifest)
	if err != nil {
		return "", false
	}
	for _, section := range dependencySections {
		if deps := doc.Object(section, false); deps != nil {
			if spec, ok := deps.String(name); ok {
				return spec, true
			}
		}
	}
	return "", false
}
