package rubygems

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Spec is what a platform gem's metadata says about it.
type Spec struct {
	Name, Version, Platform string
	Summary, Description    string
	Authors                 []string
	License, Homepage       string
	// SourceCode is the repository's URL, shown on rubygems.org.
	SourceCode string
	// Files are the gem's paths; Executables the commands in its exe/.
	Files, Executables []string
}

// specDate is every gem's date, as `gem build` stamps it without
// SOURCE_DATE_EPOCH: a gem's bytes depend on its content only.
const specDate = "1980-01-02 00:00:00.000000000 Z"

// YAML is the gem's metadata: a Gem::Specification as `gem build` writes it
// (specification version 4). Strings are double-quoted, which YAML reads
// like JSON.
func (s Spec) YAML() string {
	var b strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	list := func(key string, items []string) {
		if len(items) == 0 {
			line("%s: []", key)
			return
		}
		line("%s:", key)
		for _, it := range items {
			line("- %s", quote(it))
		}
	}
	line("--- !ruby/object:Gem::Specification")
	line("name: %s", quote(s.Name))
	line("version: !ruby/object:Gem::Version")
	line("  version: %s", quote(s.Version))
	line("platform: %s", quote(s.Platform))
	list("authors", s.Authors)
	line("bindir: exe")
	line("cert_chain: []")
	line("date: %s", specDate)
	line("dependencies: []")
	if s.Description != "" {
		line("description: %s", quote(s.Description))
	}
	list("executables", s.Executables)
	line("extensions: []")
	line("extra_rdoc_files: []")
	list("files", s.Files)
	if s.Homepage != "" {
		line("homepage: %s", quote(s.Homepage))
	}
	if s.License != "" {
		list("licenses", []string{s.License})
	} else {
		line("licenses: []")
	}
	if s.SourceCode != "" || s.Homepage != "" {
		line("metadata:")
		if s.Homepage != "" {
			line("  homepage_uri: %s", quote(s.Homepage))
		}
		if s.SourceCode != "" {
			line("  source_code_uri: %s", quote(s.SourceCode))
		}
	} else {
		line("metadata: {}")
	}
	line("rdoc_options: []")
	line("require_paths:")
	line("- lib")
	for _, req := range []string{"required_ruby_version", "required_rubygems_version"} {
		line("%s: !ruby/object:Gem::Requirement", req)
		line("  requirements:")
		line(`  - - ">="`)
		line("    - !ruby/object:Gem::Version")
		line("      version: '0'")
	}
	line("requirements: []")
	line("rubygems_version: 3.6.9")
	line("specification_version: 4")
	line("summary: %s", quote(s.Summary))
	line("test_files: []")
	return b.String()
}

// quote is a YAML double-quoted scalar; JSON's escapes are YAML's.
func quote(s string) string {
	q, _ := json.Marshal(s)
	return string(q)
}
