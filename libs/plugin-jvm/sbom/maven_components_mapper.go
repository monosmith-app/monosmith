package sbom

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func maven(group, artifact, version string) protocol.Component {
	c := components.New("maven", group, artifact, version)
	return c
}

// ComponentsFromGradleLock lists the dependencies a gradle.lockfile pins
// (group:artifact:version=configurations). One that only the test
// configurations resolve is not shipped.
func ComponentsFromGradleLock(lock string) []protocol.Component {
	var out []protocol.Component
	for _, line := range strings.Split(lock, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "empty=") {
			continue
		}
		coords, configs, _ := strings.Cut(line, "=")
		parts := strings.Split(coords, ":")
		if len(parts) != 3 || onlyTest(configs) {
			continue
		}
		out = append(out, maven(parts[0], parts[1], parts[2]))
	}
	return components.Dedupe(out)
}

func onlyTest(configs string) bool {
	if configs == "" {
		return false
	}
	for _, c := range strings.Split(configs, ",") {
		if !strings.HasPrefix(c, "test") {
			return false
		}
	}
	return true
}

var gradleDep = regexp.MustCompile(`(?m)^\s*(?:api|implementation|compile|runtimeOnly|compileOnly)\s*\(?\s*["']([^:"'\s]+):([^:"'\s]+):([^:"'\s]+)["']`)

// ComponentsFromGradleScript lists the group:artifact:version coordinates
// a build.gradle or build.gradle.kts declares as shipped dependencies.
func ComponentsFromGradleScript(script string) []protocol.Component {
	var out []protocol.Component
	for _, m := range gradleDep.FindAllStringSubmatch(script, -1) {
		out = append(out, maven(m[1], m[2], m[3]))
	}
	return components.Dedupe(out)
}

// ComponentsFromPOM lists the compile and runtime dependencies of a pom.xml.
// A ${property} version is resolved from the pom's properties, else left empty.
func ComponentsFromPOM(pom []byte) ([]protocol.Component, error) {
	dec := xml.NewDecoder(bytes.NewReader(pom))
	props := map[string]string{}
	type dep struct{ group, artifact, version, scope string }
	var deps []dep
	var path []string
	var cur *dep
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			path = append(path, t.Name.Local)
			if t.Name.Local == "dependency" && inDependencies(path) {
				cur = &dep{}
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text == "" || len(path) == 0 {
				continue
			}
			leaf := path[len(path)-1]
			switch {
			case cur != nil && len(path) >= 2 && path[len(path)-2] == "dependency":
				switch leaf {
				case "groupId":
					cur.group = text
				case "artifactId":
					cur.artifact = text
				case "version":
					cur.version = text
				case "scope":
					cur.scope = text
				}
			case len(path) == 3 && path[1] == "properties":
				props[leaf] = text
			}
		case xml.EndElement:
			if t.Name.Local == "dependency" && cur != nil {
				deps = append(deps, *cur)
				cur = nil
			}
			path = path[:len(path)-1]
		}
	}
	var out []protocol.Component
	for _, d := range deps {
		if d.scope == "test" || d.scope == "provided" || d.group == "" || d.artifact == "" {
			continue
		}
		v := d.version
		if name, ok := strings.CutPrefix(strings.TrimSuffix(v, "}"), "${"); ok {
			v = props[name]
		}
		out = append(out, maven(d.group, d.artifact, v))
	}
	return components.Dedupe(out), nil
}

// inDependencies is true for project/dependencies/dependency, not for the
// ones under dependencyManagement or a plugin.
func inDependencies(path []string) bool {
	return len(path) == 3 && path[1] == "dependencies"
}
