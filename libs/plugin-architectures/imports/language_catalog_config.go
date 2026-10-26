package imports

import "sort"

var all = []Language{goLanguage, pythonLanguage, nodeLanguage, jvmLanguage, dotnetLanguage, phpLanguage, dartLanguage, rustLanguage, rubyLanguage, cppLanguage}

// aliases are other names a project's language goes by.
var aliases = map[string]string{"javascript": "node", "js": "node", "typescript": "node", "ts": "node"}

// For returns a language's way of reading imports.
func For(language string) (Language, bool) {
	if canonical, ok := aliases[language]; ok {
		language = canonical
	}
	for _, l := range all {
		if l.Name == language {
			return l, true
		}
	}
	return Language{}, false
}

// Names lists the languages the check reads, sorted.
func Names() []string {
	var names []string
	for _, l := range all {
		names = append(names, l.Name)
	}
	sort.Strings(names)
	return names
}
