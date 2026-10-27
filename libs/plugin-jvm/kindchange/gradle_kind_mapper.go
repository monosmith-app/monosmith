package kindchange

import (
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var kindProperty = regexp.MustCompile(`(?m)^(\s*smith\.kind\s*[=:]\s*)(\S+)`)

// changeGradle writes a Gradle module's kind (smith.kind in the
// gradle.properties beside its script) and says what its script still needs:
// the build script is code, so smith does not rewrite its plugins.
func changeGradle(edits *fileedits.Set, params protocol.ChangeKindParams, result *protocol.ChangeKindResult) error {
	p := params.Project
	propertiesFile := path.Join(p.Root, "gradle.properties")
	properties, _ := edits.Read(propertiesFile)
	if kindProperty.MatchString(properties) {
		properties = kindProperty.ReplaceAllString(properties, "${1}"+string(params.To))
	} else {
		if properties != "" && !strings.HasSuffix(properties, "\n") {
			properties += "\n"
		}
		properties += "smith.kind=" + string(params.To) + "\n"
	}
	edits.Create(propertiesFile, properties)

	script := ""
	for _, name := range []string{"build.gradle.kts", "build.gradle"} {
		if text, ok := edits.Read(path.Join(p.Root, name)); ok {
			script = text
			break
		}
	}
	application := strings.Contains(script, "application")
	publishing := strings.Contains(script, "maven-publish")
	switch {
	case params.To == protocol.ProjectApp && !application:
		warnf(result, "an app runs a program: add the application plugin and its mainClass to the build script of %s", p.Root)
	case params.To != protocol.ProjectApp && application:
		warnf(result, "this module's script applies the application plugin; as a %s it is a dependency of other modules, and the program is kept but not deployed", params.To)
	}
	if params.To == protocol.ProjectPackage && !publishing {
		warnf(result, "a package is published: add the maven-publish plugin and a publishing block to the build script of %s", p.Root)
	}
	return nil
}
