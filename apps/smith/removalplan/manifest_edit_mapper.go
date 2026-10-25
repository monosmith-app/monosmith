package removalplan

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// removableSections are the package.json sections a removal takes a package
// out of; peerDependencies is a package's published contract, never edited.
var removableSections = []string{"dependencies", "devDependencies", "optionalDependencies"}

// RemovePackages returns a package.json without the named packages: out of
// its dependency sections, its npm overrides, and a section left empty.
func RemovePackages(text string, names []string) (string, error) {
	doc, err := jsondoc.ParseDocument([]byte(text))
	if err != nil {
		return "", err
	}
	for _, section := range removableSections {
		deps := doc.Object(section, false)
		if deps == nil {
			continue
		}
		for _, name := range names {
			deps.Delete(name)
		}
		if deps.Len() == 0 {
			doc.Delete(section)
		}
	}
	if overrides := doc.Object("overrides", false); overrides != nil {
		for _, name := range names {
			overrides.Delete(name)
		}
		if overrides.Len() == 0 {
			doc.Delete("overrides")
		}
	}
	return string(doc.Bytes()), nil
}

// AddDevDependency returns a package.json with the package in its
// devDependencies, unless it already depends on it in any section.
func AddDevDependency(text, name, versionRange string) (string, error) {
	doc, err := jsondoc.ParseDocument([]byte(text))
	if err != nil {
		return "", err
	}
	for _, section := range append([]string{"peerDependencies"}, removableSections...) {
		if deps := doc.Object(section, false); deps != nil && deps.Has(name) {
			return text, nil
		}
	}
	doc.Object("devDependencies", true).SetString(name, versionRange)
	return string(doc.Bytes()), nil
}
